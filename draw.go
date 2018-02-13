// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"image"
	"image/draw"
	"math"
)

type drawInput struct {
	workerOutput
	xposition int
	yposition int
	rows      int
	columns   int
}

func (g *Generator) drawSprite(opts GenSpriteOptions, imgs <-chan workerOutput, workersErrs <-chan error, inputErrs <-chan error) (image.Image, error) {
	var sprite *image.RGBA

	columns := int(opts.Columns)
	if n := opts.N(); columns > n {
		columns = n
	}
	rows := int(math.Ceil(float64(opts.N()) / float64(columns)))

	for {
		select {
		case output, ok := <-imgs:
			if !ok {
				return sprite, nil
			}
			pos := int((output.timecode - opts.Start) / opts.Interval)
			ypos := pos / columns
			xpos := pos - ypos*columns
			input := drawInput{
				workerOutput: output,
				xposition:    xpos,
				yposition:    ypos,
				rows:         rows,
				columns:      columns,
			}
			if sprite == nil {
				sprite = g.initSprite(opts, input)
			}
			g.draw(sprite, input)
		case err := <-workersErrs:
			return nil, err
		case err := <-inputErrs:
			return nil, err
		case <-opts.Context.Done():
			return nil, opts.Context.Err()
		}
	}
}

func (g *Generator) initSprite(opts GenSpriteOptions, input drawInput) *image.RGBA {
	width := input.img.Bounds().Dx()
	height := input.img.Bounds().Dy()
	spriteRect := image.Rect(0, 0, width*input.columns, height*input.rows)
	return image.NewRGBA(spriteRect)
}

func (g *Generator) draw(sprite draw.Image, input drawInput) {
	width := input.img.Bounds().Dx()
	height := input.img.Bounds().Dy()
	sp := image.Pt(width*input.xposition, height*input.yposition)
	r := image.Rectangle{sp, sp.Add(image.Pt(width, height))}
	draw.Draw(sprite, r, input.img, image.Pt(0, 0), draw.Src)
}
