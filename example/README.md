# nyt-devito

This folder presents an example of vod-module-sprite in conjunction with
[NYTimes' sample
configuration](https://github.com/NYTimes/nginx-vod-module-docker/tree/HEAD/examples)
for nginx-vod-module.

It provides a command line utility that uses vod-module-sprite to generate
sprites and save them to a file in the local disk.

## Using this example

Make sure you have nginx-vod-module running on port 3030, as documented in
nginx-vod-module-docker example:
https://github.com/NYTimes/nginx-vod-module-docker/tree/HEAD/examples

Once nginx-vod-module is running, compile and run
[nyt-devito.go](/example/nyt-devito.go):

```
% go build -o nyt-devito nyt-devito.go
% ./nyt-devito
2018/02/10 14:28:13 successfully generated thumbnail "thumb.jpg"
```

There are some flags that allow customizing the behavior of the tool:

```
% ./nyt-devito -h
Usage of ./nyt-devito:
  -end duration
    	timecode for the end point (default 2m0s)
  -height uint
    	height of each sprite item - 0 for keeping the aspect ratio/source
  -interval duration
    	interval between captures (default 2s)
  -keep-ratio
    	keep aspect ratio?
  -max-workers uint
    	maximum number of workers to be used for thumbnail generation (default 32)
  -o string
    	output file (default "thumb.jpg")
  -packager string
    	endpoint of the packager (default "http://localhost:3030")
  -start duration
    	timecode for the starting point
  -url string
    	url of the source video (default "http://localhost:3030/videos/devito480p.mp4")
  -width uint
    	width of each sprite item - 0 for keeping the aspect ratio/source
```
