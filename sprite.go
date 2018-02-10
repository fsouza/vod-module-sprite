package sprite

import (
	"bytes"
	"errors"
	"image/jpeg"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

var mp4Regexp = regexp.MustCompile(`\.mp4$`)

// Generator generates sprites for videos using the video-packager.
type Generator struct {
	VideoPackagerEndpoint string
	MaxWorkers            uint

	client *http.Client
	o      sync.Once
}

// GenSpriteOptions is the set of options that control the sprite generation
// for a video rendition.
type GenSpriteOptions struct {
	RenditionURL string
	Start        time.Duration
	End          time.Duration
	Interval     time.Duration
	Width        uint
	Height       uint
	JPEGQuality  int

	prefix string
}

// N returns the number of items expected to be present in the generated
// sprite.
func (o *GenSpriteOptions) N() int {
	return int((o.End-o.Start)/o.Interval) + 1
}

// GenSprite generates the sprite for the given video.
//
// It takes the rendition URL, the duration and the interval.
func (g *Generator) GenSprite(opts GenSpriteOptions) ([]byte, error) {
	g.initGenerator()
	prefix, err := g.videoPackagerPrefix(opts.RenditionURL)
	if err != nil {
		return nil, err
	}
	opts.prefix = prefix
	var wg sync.WaitGroup
	inputs, abort, imgs, errs := g.startWorkers(&wg)

	err = g.sendInputs(opts, inputs, errs)
	if err != nil {
		close(abort)
		wg.Wait()
		return nil, err
	}

	sprite, err := g.drawSprite(opts, imgs, errs)
	if err != nil {
		close(abort)
		wg.Wait()
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, sprite, &jpeg.Options{Quality: opts.JPEGQuality})
	return buf.Bytes(), err
}

func (g *Generator) initGenerator() {
	g.o.Do(func() {
		g.client = cleanhttp.DefaultPooledClient()
	})
}

func (g *Generator) startWorkers(wg *sync.WaitGroup) (chan<- workerInput, chan<- struct{}, <-chan workerOutput, <-chan error) {
	nworkers := int(g.MaxWorkers)
	inputs := make(chan workerInput, nworkers)
	imgs := make(chan workerOutput, nworkers)
	errs := make(chan error, nworkers+1)
	abort := make(chan struct{})
	for i := 0; i < nworkers; i++ {
		wg.Add(1)
		w := worker{client: g.client, group: wg}
		go w.Run(inputs, abort, imgs, errs)
	}
	go func() {
		wg.Wait()
		close(imgs)
	}()
	return inputs, abort, imgs, errs
}

func (g *Generator) sendInputs(opts GenSpriteOptions, inputs chan<- workerInput, errs <-chan error) error {
	defer close(inputs)
	for timecode := opts.Start; timecode <= opts.End; timecode += opts.Interval {
		input := workerInput{
			prefix:   opts.prefix,
			width:    opts.Width,
			height:   opts.Height,
			timecode: timecode,
		}

		select {
		case inputs <- input:
		case err := <-errs:
			return err
		}
	}
	return nil
}

func (g *Generator) videoPackagerPrefix(renditionURL string) (string, error) {
	rurl, err := url.Parse(renditionURL)
	if err != nil {
		return "", err
	}
	path := rurl.Path
	if !strings.HasPrefix(path, "/video/") {
		return "", errors.New("invalid rendition: path doesn't start with /video/")
	}
	if !mp4Regexp.MatchString(path) {
		return "", errors.New("invalid rendition: not an mp4 file")
	}
	path = strings.Replace(path, "/video/", "/video/t/", 1)
	path = mp4Regexp.ReplaceAllString(path, "/")
	return strings.TrimRight(g.VideoPackagerEndpoint, "/") + path, nil
}
