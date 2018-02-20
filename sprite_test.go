// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenSprite(t *testing.T) {
	const (
		testJPEGQuality = 100
		maxDiff         = int64(11e5)
	)

	var tests = []struct {
		name         string
		input        GenSpriteOptions
		expectedFile string
	}{
		{
			"full sprite - vertical",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-full.jpg",
		},
		{
			"full sprite - vertical with black bars",
			GenSpriteOptions{
				VideoURL:        "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:           0,
				End:             18 * time.Second,
				Interval:        2 * time.Second,
				Width:           200,
				Height:          72,
				JPEGQuality:     testJPEGQuality,
				KeepAspectRatio: true,
			},
			"sprite-full-blackbars.jpg",
		},
		{
			"full sprite - horizontal",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     10,
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-full-horizontal.jpg",
		},
		{
			"full sprite - horizontal with black bars",
			GenSpriteOptions{
				VideoURL:        "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:         10,
				Start:           0,
				End:             18 * time.Second,
				Interval:        2 * time.Second,
				Width:           200,
				Height:          72,
				JPEGQuality:     testJPEGQuality,
				KeepAspectRatio: true,
			},
			"sprite-full-blackbars-horizontal.jpg",
		},
		{
			"full sprite - 2 columns",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     2,
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-full-2-columns.jpg",
		},
		{
			"partial sprite - just the start",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       0,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-0-14000.jpg",
		},
		{
			"partial sprite - just the end",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       4 * time.Second,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-4000.jpg",
		},
		{
			"partial sprite - middle - vertical",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       4 * time.Second,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-4000-14000.jpg",
		},
		{
			"partial sprite - middle - horizontal",
			GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     10,
				Start:       4 * time.Second,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			"sprite-4000-14000-horizontal.jpg",
		},
	}

	const spritesFolder = "testdata"
	packager := startFakePackager(spritesFolder)
	defer packager.stop()
	generator := Generator{Translator: packager.translate, MaxWorkers: 4}

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
				if f, err := ioutil.TempFile("", ""); err == nil {
					f.Write(data)
					f.Close()
					t.Logf("output sprite saved to: %v", f.Name())
				}
				t.Errorf("images are too different\nmax diff: %d\ngot diff: %d", maxDiff, diff)
				t.Errorf("expectedSprite: %v", test.expectedFile)
			}
		})
	}
}

func TestGenSpriteErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	var tests = []struct {
		name  string
		input GenSpriteOptions
	}{
		{
			"invalid timecodes",
			GenSpriteOptions{
				VideoURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    30 * time.Second,
				End:      50 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
		},
		{
			"invalid rendition",
			GenSpriteOptions{
				VideoURL: "/videos/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    4 * time.Second,
				End:      14 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
		},
		{
			"context cancelation",
			GenSpriteOptions{
				Context:  ctx,
				VideoURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    4 * time.Second,
				End:      14 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
		},
	}

	const spritesFolder = "testdata"
	packager := startFakePackager(spritesFolder)
	defer packager.stop()
	generator := Generator{Translator: packager.translate, MaxWorkers: 32}

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
