package sprite

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"net/http"
	"net/url"
	"regexp"
	"sort"
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
	Duration     time.Duration
	Interval     time.Duration
	Width        uint
	Height       uint
	JPEGQuality  int

	prefix string
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

	outputs, err := g.collectOutputs(imgs, errs)
	if err != nil {
		close(abort)
		wg.Wait()
		return nil, err
	}
	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].timecode < outputs[j].timecode
	})
	return g.drawSprite(opts, outputs)
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
	for timecode := time.Duration(0); timecode < opts.Duration; timecode += opts.Interval {
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

func (g *Generator) collectOutputs(imgs <-chan workerOutput, errs <-chan error) ([]workerOutput, error) {
	var outputs []workerOutput

	for {
		select {
		case output, ok := <-imgs:
			if !ok {
				return outputs, nil
			}
			outputs = append(outputs, output)
		case err := <-errs:
			return nil, err
		}
	}
}

func (g *Generator) drawSprite(opts GenSpriteOptions, outputs []workerOutput) ([]byte, error) {
	if len(outputs) < 1 {
		return nil, nil
	}

	sample := outputs[0].img
	width := sample.Bounds().Dx()
	height := sample.Bounds().Dy()
	spriteRect := image.Rect(0, 0, width, height*len(outputs))
	sprite := image.NewRGBA(spriteRect)

	sp := image.Pt(0, 0)
	r := image.Rect(0, 0, width, height)
	for _, output := range outputs {
		draw.Draw(sprite, r, output.img, image.Point{0, 0}, draw.Src)
		sp = sp.Add(image.Pt(0, height))
		r = image.Rectangle{sp, sp.Add(image.Pt(width, height))}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, sprite, &jpeg.Options{Quality: opts.JPEGQuality})
	return buf.Bytes(), err
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
