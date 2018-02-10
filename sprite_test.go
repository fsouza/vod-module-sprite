package sprite

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenSprite(t *testing.T) {
	const (
		testJPEGQuality = 80
		maxDiff         = int64(1e6)
	)

	var tests = []struct {
		name         string
		input        GenSpriteOptions
		expectedFile string
	}{
		{
			"full sprite",
			GenSpriteOptions{
				RenditionURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        0,
				End:          18 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
				JPEGQuality:  testJPEGQuality,
			},
			"sprite-full.jpg",
		},
		{
			"partial sprite - just the start",
			GenSpriteOptions{
				RenditionURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        0,
				End:          14 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
				JPEGQuality:  testJPEGQuality,
			},
			"sprite-0-14000.jpg",
		},
		{
			"partial sprite - just the end",
			GenSpriteOptions{
				RenditionURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        4 * time.Second,
				End:          18 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
				JPEGQuality:  testJPEGQuality,
			},
			"sprite-4000.jpg",
		},
		{
			"partial sprite - middle",
			GenSpriteOptions{
				RenditionURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        4 * time.Second,
				End:          14 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
				JPEGQuality:  testJPEGQuality,
			},
			"sprite-4000-14000.jpg",
		},
	}

	const spritesFolder = "testdata"
	packager := startFakePackager(spritesFolder)
	defer packager.stop()
	generator := Generator{VideoPackagerEndpoint: packager.url(), MaxWorkers: 32}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := generator.GenSprite(test.input)
			if err != nil {
				t.Fatal(err)
			}
			sprite, err := jpeg.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("GenSprite didn't generate a valid jpeg: %v", err)
			}
			expectedSprite, err := loadSpriteFromDisk(filepath.Join(spritesFolder, test.expectedFile))
			if err != nil {
				t.Fatal(err)
			}
			if sprite.Bounds() != expectedSprite.Bounds() {
				t.Errorf("image bounds don't match\nwant %v\ngot  %v", expectedSprite.Bounds(), sprite.Bounds())
			}
			diff := imageDiff(sprite, expectedSprite)
			if int64(math.Abs(float64(diff))) > maxDiff {
				t.Errorf("images are too different\nmax diff: %d\ngot diff: %d", maxDiff, diff)
			}
		})
	}
}

func TestGenSpriteErrors(t *testing.T) {
	var tests = []struct {
		name  string
		input GenSpriteOptions
	}{
		{
			"invalid timecodes",
			GenSpriteOptions{
				RenditionURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        30 * time.Second,
				End:          50 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
			},
		},
		{
			"invalid rendition",
			GenSpriteOptions{
				RenditionURL: "/videos/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:        4 * time.Second,
				End:          14 * time.Second,
				Interval:     2 * time.Second,
				Height:       72,
			},
		},
	}

	const spritesFolder = "testdata"
	packager := startFakePackager(spritesFolder)
	defer packager.stop()
	generator := Generator{VideoPackagerEndpoint: packager.url(), MaxWorkers: 32}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := generator.GenSprite(test.input)
			if data != nil {
				t.Error("got unexpected non-nil data")
			}
			if err == nil {
				t.Error("got unexpected <nil> error")
			}
		})
	}
}

func TestGenSpriteOptionsN(t *testing.T) {
	var tests = []struct {
		name     string
		input    GenSpriteOptions
		expected int
	}{
		{
			"0 to 18, every 2 seconds",
			GenSpriteOptions{
				Interval: 2 * time.Second,
				End:      18 * time.Second,
			},
			10,
		},
		{
			"0 to 18, every second",
			GenSpriteOptions{
				Interval: time.Second,
				End:      18 * time.Second,
			},
			19,
		},
		{
			"1 to 18, every 2 seconds",
			GenSpriteOptions{
				Start:    time.Second,
				Interval: 2 * time.Second,
				End:      18 * time.Second,
			},
			9,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n := test.input.N()
			if n != test.expected {
				t.Errorf("wrong value\nwant %d\ngot  %d", test.expected, n)
			}
		})
	}
}

func TestVideoPackagerPrefix(t *testing.T) {
	var tests = []struct {
		name      string
		endpoint  string
		rendition string
		expected  string
		errMsg    string
	}{
		{
			"regular url",
			"https://video-packager.example.com",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"extra slashes on endpoint",
			"https://video-packager.example.com////",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"rendition path",
			"https://video-packager.example.com",
			"/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"invalid rendition url - not /video/ prefix",
			"https://video-packager.example.com",
			"https://vp.nyt.com/videos/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"",
			"invalid rendition: path doesn't start with /video/",
		},
		{
			"invalid rendition url - not mp4",
			"https://video-packager.example.com",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_480p.webm",
			"",
			"invalid rendition: not an mp4 file",
		},
		{
			"invalid url",
			"https://video-packager.example.com",
			":video",
			"",
			"parse :video: missing protocol scheme",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := Generator{VideoPackagerEndpoint: test.endpoint}
			prefix, err := generator.videoPackagerPrefix(test.rendition)
			if err == nil {
				err = errors.New("")
			}
			if err.Error() != test.errMsg {
				t.Errorf("wrong error\nwant %q\ngot  %q", test.errMsg, err.Error())
			}
			if prefix != test.expected {
				t.Errorf("wrong prefix\nwant %q\ngot  %q", test.expected, prefix)
			}
		})
	}
}

// imageDiff calculates the distance between two images.
//
// The function assumes that both images have the same bounds.
//
// Source: https://stackoverflow.com/a/36439876
func imageDiff(img1, img2 image.Image) int64 {
	var accumError int64

	bounds := img1.Bounds()
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()
			accumError += sqDiff(r1, r2)
			accumError += sqDiff(g1, g2)
			accumError += sqDiff(b1, b2)
			accumError += sqDiff(a1, a2)
		}
	}

	return int64(math.Sqrt(float64(accumError)))
}

func sqDiff(x, y uint32) int64 {
	d := int64(x) - int64(y)
	return d * d
}

// loadSpriteFromDisk returns the image or an error if the file doesn't exist
// or isn't a JPEG.
func loadSpriteFromDisk(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jpeg.Decode(f)
}
