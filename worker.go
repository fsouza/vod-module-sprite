package sprite

import (
	"fmt"
	"image"
	"strconv"
	"strings"
	"sync"
	"time"
)

type workerInput struct {
	prefix   string
	timecode time.Duration
	width    uint
	height   uint
}

func (i *workerInput) url() string {
	milli := i.timecode.Truncate(time.Millisecond)
	suffixParts := []string{"thumb", strconv.FormatInt(int64(milli/time.Millisecond), 10)}
	if i.width > 0 {
		suffixParts = append(suffixParts, fmt.Sprintf("w%d", i.width))
	}
	if i.height > 0 {
		suffixParts = append(suffixParts, fmt.Sprintf("h%d", i.height))
	}
	return fmt.Sprintf("%s%s.jpg", i.prefix, strings.Join(suffixParts, "-"))
}

func worker(inputs <-chan workerInput, imgs chan<- image.Image, errs chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
}
