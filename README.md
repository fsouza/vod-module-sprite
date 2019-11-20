# vod-module-sprite

[![Build Status](https://github.com/fsouza/vod-module-sprite/workflows/Build/badge.svg)](https://github.com/fsouza/vod-module-sprite/actions?query=branch:master+workflow:Build)

Library for generating sprites from
[nginx-vod-module](https://github.com/kaltura/nginx-vod-module) thumbnails. It
leverages nginx-vod-module's thumbnail generation capabilities and stitches
thumbs together to generate a vertical sprite.

## Example

Check the [example folder](/example) for an example of sprite generation integrated with
[NYTimes' nginx-vod-module-docker sample
config](https://github.com/NYTimes/nginx-vod-module-docker/tree/master/examples).
