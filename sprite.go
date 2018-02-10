// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"bytes"
	"image/jpeg"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
)

// Generator generates sprites for videos using the video-packager.
type Generator struct {
	Translator VideoURLTranslator
	MaxWorkers uint

	client *http.Client
	o      sync.Once
}

// VideoURLTranslator is a function that translates a video URL into a
// nginx-vod-module thumb prefix URL.
//
// A thumb prefix URL is a URL that doesn't include the suffix
// `thumb-{timecode}-w{width}-h{height}`.
//
// The Generator will use this function to derive the final URL of the
// thumbnail asset.
type VideoURLTranslator func(string) (string, error)

// GenSpriteOptions is the set of options that control the sprite generation
// for a video rendition.
type GenSpriteOptions struct {
	VideoURL    string
	Start       time.Duration
	End         time.Duration
	Interval    time.Duration
	Width       uint
	Height      uint
	JPEGQuality int

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
	prefix, err := g.Translator(opts.VideoURL)
	if err != nil {
		return nil, err
	}
	opts.prefix = prefix
	var wg sync.WaitGroup
	inputs, abort, imgs, errs := g.startWorkers(opts, &wg)

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

func (g *Generator) startWorkers(opts GenSpriteOptions, wg *sync.WaitGroup) (chan<- workerInput, chan<- struct{}, <-chan workerOutput, <-chan error) {
	nworkers := opts.N()/2 + 1
	if nworkers > int(g.MaxWorkers) {
		nworkers = int(g.MaxWorkers)
	}
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
