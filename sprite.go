package sprite

import (
	"errors"
	"image"
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
	Duration     time.Duration
	Interval     time.Duration
	Width        uint
	Height       uint
}

// GenSprite generates the sprite for the given video.
//
// It takes the rendition URL, the duration and the interval.
func (g *Generator) GenSprite(opts GenSpriteOptions) ([]byte, error) {
	g.initGenerator()
	return nil, nil
}

func (g *Generator) initGenerator() {
	g.o.Do(func() {
		g.client = cleanhttp.DefaultPooledClient()
	})
}

func (g *Generator) startWorkers(wg *sync.WaitGroup) (chan<- workerInput, <-chan image.Image, <-chan error) {
	nworkers := int(g.MaxWorkers)
	inputs := make(chan workerInput, nworkers)
	imgs := make(chan image.Image, nworkers)
	errs := make(chan error, nworkers+1)
	for i := 0; i < nworkers; i++ {
		wg.Add(1)
		w := worker{client: g.client, group: wg}
		go w.Run(inputs, imgs, errs)
	}
	go func() {
		wg.Wait()
		close(imgs)
		close(errs)
	}()
	return inputs, imgs, errs
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
