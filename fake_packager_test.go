// Copyright 2018 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sprite

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type fakePackager struct {
	server         *httptest.Server
	folder         string
	router         *mux.Router
	prefixRegexp   *regexp.Regexp
	suffixRegexp   *regexp.Regexp
	o              sync.Once
	failAtTimecode []int64
}

func startFakePackager(folder string) *fakePackager {
	p := fakePackager{
		folder:       folder,
		prefixRegexp: regexp.MustCompile(`^/video/`),
		suffixRegexp: regexp.MustCompile(`\.mp4$`),
	}
	p.server = httptest.NewServer(&p)
	return &p
}

func (p *fakePackager) stop() {
	p.server.Close()
}

func (p *fakePackager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.o.Do(p.initRouter)
	p.router.ServeHTTP(w, r)
}

func (p *fakePackager) initRouter() {
	p.router = mux.NewRouter()
	p.router.HandleFunc(`/thumbs/{rendition:.+}/thumb-{timecode:\d+}.jpg`, p.genImage)
	p.router.HandleFunc(`/thumbs/{rendition:.+}/thumb-{timecode:\d+}-h{height}.jpg`, p.genImage)
	p.router.HandleFunc(`/thumbs/{rendition:.+}/thumb-{timecode:\d+}-w{width}-h{height}.jpg`, p.genImage)
}

func (p *fakePackager) translate(videoURL string) (string, error) {
	vurl, err := url.Parse(videoURL)
	if err != nil {
		return "", err
	}
	path := p.prefixRegexp.ReplaceAllString(vurl.Path, "/thumbs/")
	path = p.suffixRegexp.ReplaceAllString(path, "")
	if path == vurl.Path {
		//nolint:goerr113
		return "", errors.New("invalid videoURL")
	}
	return p.server.URL + path, nil
}

func (p *fakePackager) genImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timecode, _ := strconv.ParseInt(vars["timecode"], 10, 64)
	if p.shouldFail(timecode) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	fileName := p.fileName(timecode)
	if fileName == "" {
		http.Error(w, "invalid timecode", http.StatusBadRequest)
		return
	}
	f, err := os.Open(filepath.Join(p.folder, fileName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, f)
}

func (p *fakePackager) shouldFail(timecode int64) bool {
	for _, t := range p.failAtTimecode {
		if t == timecode {
			return true
		}
	}
	return false
}

func (p *fakePackager) fileName(timecode int64) string {
	files := map[int64]string{
		0:     "img01.jpg",
		2000:  "img02.jpg",
		4000:  "img03.jpg",
		6000:  "img04.jpg",
		8000:  "img05.jpg",
		10000: "img06.jpg",
		12000: "img07.jpg",
		14000: "img08.jpg",
		16000: "img09.jpg",
		18000: "img10.jpg",
	}
	return files[timecode]
}
