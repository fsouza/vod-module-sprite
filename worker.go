// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VideoPackagerError represents an error reported by the video packager.
type VideoPackagerError struct {
	StatusCode   int
	ResponseBody []byte
}

// Error returns the string representation of VideoPackagerError.
func (err *VideoPackagerError) Error() string {
	return fmt.Sprintf("invalid response from video-packager: %d - %s", err.StatusCode, err.ResponseBody)
}

type workerInput struct {
	prefix       string
	timecode     time.Duration
	width        uint
	height       uint
	addBlackBars bool
}

func (i *workerInput) url() string {
	milli := i.timecode.Truncate(time.Millisecond)
	suffixParts := []string{"thumb", strconv.FormatInt(int64(milli/time.Millisecond), 10)}
	if i.width > 0 && !i.addBlackBars {
		suffixParts = append(suffixParts, fmt.Sprintf("w%d", i.width))
	}
	if i.height > 0 {
		suffixParts = append(suffixParts, fmt.Sprintf("h%d", i.height))
	}
	return fmt.Sprintf("%s/%s.jpg", strings.TrimRight(i.prefix, "/"), strings.Join(suffixParts, "-"))
}

type workerOutput struct {
	img   image.Image
	input workerInput
}

type worker struct {
	client *http.Client
	group  *sync.WaitGroup
}

func (w *worker) Run(ctx context.Context, inputs <-chan workerInput, abort <-chan struct{}, imgs chan<- workerOutput, errs chan<- error) {
	defer w.group.Done()
	for {
		select {
		case input, ok := <-inputs:
			if !ok {
				return
			}

			img, err := w.process(input)
			if err != nil {
				errs <- err
				return
			}

			select {
			case imgs <- workerOutput{img: img, input: input}:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			case <-abort:
				return
			}
		case <-ctx.Done():
			errs <- ctx.Err()
			return
		case <-abort:
			return
		}
	}
}

func (w *worker) process(input workerInput) (image.Image, error) {
	thumbURL := input.url()
	resp, err := w.client.Get(thumbURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, &VideoPackagerError{
			StatusCode:   resp.StatusCode,
			ResponseBody: data,
		}
	}
	return jpeg.Decode(resp.Body)
}
