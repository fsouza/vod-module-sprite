# vod-module-sprite

[![Build Status](https://travis-ci.org/fsouza/vod-module-sprite.svg?branch=master)](https://travis-ci.org/fsouza/vod-module-sprite)
[![codecov](https://codecov.io/gh/fsouza/vod-module-sprite/branch/master/graph/badge.svg)](https://codecov.io/gh/fsouza/vod-module-sprite)

Library for generating sprites from
[nginx-vod-module](https://github.com/kaltura/nginx-vod-module) thumbnails. It
leverages nginx-vod-module's thumbnail generation capabilities and stitch the
thumbs together to generate a vertical sprite.

## Example

Check the [example folder](/example) for an example of sprite generation integrated with
[NYTimes' nginx-vod-module-docker sample
config](https://github.com/NYTimes/nginx-vod-module-docker/tree/master/examples).
