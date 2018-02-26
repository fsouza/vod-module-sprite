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

func (i *drawInput) dimensions() (width, height int) {
	width = i.img.Bounds().Dx()
	height = i.img.Bounds().Dy()
	if i.workerOutput.input.addBlackBars {
		width = int(i.workerOutput.input.width)
	}
	return width, height
}

func (g *Generator) drawSprite(opts GenSpriteOptions, imgs <-chan workerOutput, workersErrs <-chan error, inputErrs <-chan error) (image.Image, error) {
	var drawer spriteDrawer

	columns := int(opts.Columns)
	if n := opts.n(); columns > n {
		columns = n
	}
	rows := int(math.Ceil(float64(opts.n()) / float64(columns)))

	for {
		select {
		case output, ok := <-imgs:
			if !ok {
				return drawer.sprite, nil
			}
			pos := int((output.input.timecode - opts.Start) / opts.Interval)
			ypos := pos / columns
			xpos := pos - ypos*columns
			input := drawInput{
				workerOutput: output,
				xposition:    xpos,
				yposition:    ypos,
				rows:         rows,
				columns:      columns,
			}
			drawer.draw(input)
		case err := <-workersErrs:
			return nil, err
		case err := <-inputErrs:
			return nil, err
		case <-opts.Context.Done():
			return nil, opts.Context.Err()
		}
	}
}

type spriteDrawer struct {
	sprite *image.RGBA
}

func (d *spriteDrawer) draw(input drawInput) {
	width, height := input.dimensions()
	if d.sprite == nil {
		spriteRect := image.Rect(0, 0, width*input.columns, height*input.rows)
		d.sprite = image.NewRGBA(spriteRect)
	}

	var offset int
	if input.workerOutput.input.addBlackBars {
		if diff := width - input.img.Bounds().Dx(); diff > 0 {
			offset = diff / 2
		}
	}

	sp := image.Pt(width*input.xposition+offset, height*input.yposition)
	r := image.Rectangle{sp, sp.Add(image.Pt(width, height))}
	draw.Draw(d.sprite, r, input.img, image.Pt(0, 0), draw.Src)
}
