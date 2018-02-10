package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/vod-module-sprite"
)

func main() {
	packagerEndpoint := flag.String("packager", "http://localhost:3030", "endpoint of the packager")
	maxWorkers := flag.Uint("max-workers", 32, "maximum number of workers to be used for thumbnail generation")
	output := flag.String("o", "thumb.jpg", "output file")
	url := flag.String("url", "http://localhost:3030/videos/devito480p.mp4", "url of the source video")
	width := flag.Uint("width", 0, "width of each sprite item - 0 for keeping the aspect ratio/source")
	height := flag.Uint("height", 0, "height of each sprite item - 0 for keeping the aspect ratio/source")
	interval := flag.Duration("interval", 2*time.Second, "interval between captures")
	start := flag.Duration("start", 0, "timecode for the starting point")
	end := flag.Duration("end", 2*time.Minute, "timecode for the end point")
	flag.Parse()

	generator := sprite.Generator{
		Translator: getTranslator(*packagerEndpoint),
		MaxWorkers: *maxWorkers,
	}
	data, err := generator.GenSprite(sprite.GenSpriteOptions{
		VideoURL:    *url,
		Width:       *width,
		Height:      *height,
		Start:       *start,
		End:         *end,
		Interval:    *interval,
		JPEGQuality: 80,
	})
	if err != nil {
		log.Fatalf("failed to generate sprite: %v", err)
	}
	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	if n != len(data) {
		log.Fatalf("failed to write %q: %v", *output, io.ErrShortWrite)
	}
	log.Printf("successfully generated thumbnail %q", *output)
}

func getTranslator(packagerEndpoint string) sprite.VideoURLTranslator {
	videoMatch := regexp.MustCompile(`^/videos/(.*)$`)
	packagerEndpoint = strings.TrimRight(packagerEndpoint, "/")
	return func(videoURL string) (string, error) {
		vurl, err := url.Parse(videoURL)
		if err != nil {
			return "", err
		}
		path := videoMatch.ReplaceAllString(vurl.Path, "/thumb/$1")
		if path == vurl.Path {
			return "", errors.New("invalid video url")
		}
		return packagerEndpoint + path, nil
	}
}
