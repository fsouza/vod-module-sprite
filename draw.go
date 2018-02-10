package sprite

import (
	"image"
	"image/draw"
)

type drawInput struct {
	workerOutput
	position int
	total    int
}

func (g *Generator) drawSprite(opts GenSpriteOptions, imgs <-chan workerOutput, errs <-chan error) (image.Image, error) {
	var sprite *image.RGBA

	for {
		select {
		case output, ok := <-imgs:
			if !ok {
				return sprite, nil
			}
			input := drawInput{
				workerOutput: output,
				position:     int((output.timecode - opts.Start) / opts.Interval),
				total:        opts.N(),
			}
			if sprite == nil {
				sprite = g.initSprite(input)
			}
			g.draw(sprite, input)
		case err := <-errs:
			return nil, err
		}
	}
}

func (g *Generator) initSprite(input drawInput) *image.RGBA {
	width := input.img.Bounds().Dx()
	height := input.img.Bounds().Dy()
	spriteRect := image.Rect(0, 0, width, height*input.total)
	return image.NewRGBA(spriteRect)
}

func (g *Generator) draw(sprite *image.RGBA, input drawInput) {
	width := input.img.Bounds().Dx()
	height := input.img.Bounds().Dy()
	sp := image.Pt(0, height*input.position)
	r := image.Rectangle{sp, sp.Add(image.Pt(width, height))}
	draw.Draw(sprite, r, input.img, image.Pt(0, 0), draw.Src)
}
