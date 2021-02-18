// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"bytes"
	"context"
	"errors"
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
	t.Parallel()
	const (
		testJPEGQuality = 100
		maxDiff         = int64(11e5)
	)

	tests := []struct {
		name           string
		input          GenSpriteOptions
		failAtTimecode []int64
		expectedFile   string
	}{
		{
			name: "full sprite - vertical",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-full.jpg",
		},
		{
			name: "full sprite - vertical with black bars",
			input: GenSpriteOptions{
				VideoURL:        "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:           0,
				End:             18 * time.Second,
				Interval:        2 * time.Second,
				Width:           200,
				Height:          72,
				JPEGQuality:     testJPEGQuality,
				KeepAspectRatio: true,
			},
			expectedFile: "sprite-full-blackbars.jpg",
		},
		{
			name: "full sprite - horizontal",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     10,
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-full-horizontal.jpg",
		},
		{
			name: "full sprite - horizontal with black bars",
			input: GenSpriteOptions{
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
			expectedFile: "sprite-full-blackbars-horizontal.jpg",
		},
		{
			name: "full sprite - 2 columns",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     2,
				Start:       0,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-full-2-columns.jpg",
		},
		{
			name: "partial sprite - just the start",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       0,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-0-14000.jpg",
		},
		{
			name: "partial sprite - just the end",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       4 * time.Second,
				End:         18 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-4000.jpg",
		},
		{
			name: "partial sprite - middle - vertical",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:       4 * time.Second,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-4000-14000.jpg",
		},
		{
			name: "partial sprite - middle - horizontal",
			input: GenSpriteOptions{
				VideoURL:    "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Columns:     10,
				Start:       4 * time.Second,
				End:         14 * time.Second,
				Interval:    2 * time.Second,
				Height:      72,
				JPEGQuality: testJPEGQuality,
			},
			expectedFile: "sprite-4000-14000-horizontal.jpg",
		},
		{
			name: "full sprite - vertical - ContinueOnError",
			input: GenSpriteOptions{
				VideoURL:        "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:           0,
				End:             18 * time.Second,
				Interval:        2 * time.Second,
				Height:          72,
				JPEGQuality:     testJPEGQuality,
				ContinueOnError: true,
			},
			failAtTimecode: []int64{2000, 8000},
			expectedFile:   "sprite-full-error.jpg",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			const spritesFolder = "testdata"
			packager := startFakePackager(spritesFolder)
			defer packager.stop()
			generator := Generator{Translator: packager.translate, MaxWorkers: 4}
			packager.failAtTimecode = test.failAtTimecode

			expectedSprite, err := loadSpriteFromDisk(filepath.Join(spritesFolder, test.expectedFile))
			if err != nil {
				t.Fatal(err)
			}
			data, err := generator.GenSprite(test.input)
			if err != nil {
				t.Fatal(err)
			}
			sprite, err := jpeg.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("GenSprite didn't generate a valid jpeg: %v", err)
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
	t.Parallel()
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	tests := []struct {
		name    string
		input   GenSpriteOptions
		httpErr bool
	}{
		{
			name: "invalid timecodes",
			input: GenSpriteOptions{
				VideoURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    30 * time.Second,
				End:      50 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
			httpErr: true,
		},
		{
			name: "ignores ContinueOnErro for 400s",
			input: GenSpriteOptions{
				VideoURL:        "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:           30 * time.Second,
				End:             50 * time.Second,
				Interval:        2 * time.Second,
				Height:          72,
				ContinueOnError: true,
			},
			httpErr: true,
		},
		{
			name: "invalid rendition",
			input: GenSpriteOptions{
				VideoURL: "/videos/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    4 * time.Second,
				End:      14 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
			httpErr: true,
		},
		{
			name: "context cancelation",
			input: GenSpriteOptions{
				Context:  ctx,
				VideoURL: "/video/2017/05/26/000000_1_CREDIT-SUISSE--O-_wg_360p.mp4",
				Start:    4 * time.Second,
				End:      14 * time.Second,
				Interval: 2 * time.Second,
				Height:   72,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			const spritesFolder = "testdata"
			packager := startFakePackager(spritesFolder)
			defer packager.stop()
			generator := Generator{Translator: packager.translate, MaxWorkers: 32}
			t.Parallel()
			data, err := generator.GenSprite(test.input)
			if data != nil {
				t.Error("got unexpected non-nil data")
			}
			if err == nil {
				t.Fatal("got unexpected <nil> error")
			}
			if test.httpErr {
				var verr *VideoPackagerError
				if !errors.As(err, &verr) {
					t.Errorf("expected %#v to be VideoPackagerError, but it wasn't", err)
				}
			}
		})
	}
}

func TestGenSpriteOptionsN(t *testing.T) {
	t.Parallel()
	tests := []struct {
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
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			n := test.input.n()
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
