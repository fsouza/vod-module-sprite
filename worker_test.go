// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"testing"
	"time"
)

func TestWorkerInputURL(t *testing.T) {
	var tests = []struct {
		name     string
		input    workerInput
		expected string
	}{
		{
			"exact duration, width and height",
			workerInput{
				prefix:   "https://video-packager.example.com/video/t/something/",
				width:    128,
				height:   72,
				timecode: 2 * time.Second,
			},
			"https://video-packager.example.com/video/t/something/thumb-2000-w128-h72.jpg",
		},
		{
			"duration, width and height + black bars",
			workerInput{
				prefix:       "https://video-packager.example.com/video/t/something/",
				width:        128,
				height:       72,
				timecode:     2 * time.Second,
				addBlackBars: true,
			},
			"https://video-packager.example.com/video/t/something/thumb-2000-h72.jpg",
		},
		{
			"non-exact duration, width and height",
			workerInput{
				prefix:   "https://video-packager.example.com/video/t/something/",
				width:    96,
				height:   72,
				timecode: 2531*time.Millisecond + 2531,
			},
			"https://video-packager.example.com/video/t/something/thumb-2531-w96-h72.jpg",
		},
		{
			"no width",
			workerInput{
				prefix:   "https://video-packager.example.com/video/t/something/",
				height:   72,
				timecode: 2 * time.Second,
			},
			"https://video-packager.example.com/video/t/something/thumb-2000-h72.jpg",
		},
		{
			"no height",
			workerInput{
				prefix:   "https://video-packager.example.com/video/t/something/",
				width:    128,
				timecode: 12 * time.Second,
			},
			"https://video-packager.example.com/video/t/something/thumb-12000-w128.jpg",
		},
		{
			"no width or height",
			workerInput{
				prefix:   "https://video-packager.example.com/video/t/something/",
				timecode: 2 * time.Second,
			},
			"https://video-packager.example.com/video/t/something/thumb-2000.jpg",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := test.input.url()
			if url != test.expected {
				t.Errorf("wrong url returned\nwant %q\ngot  %q", test.expected, url)
			}
		})
	}
}
