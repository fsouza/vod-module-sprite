package sprite

import (
	"errors"
	"testing"
)

func TestVideoPackagerPrefix(t *testing.T) {
	var tests = []struct {
		name      string
		endpoint  string
		rendition string
		expected  string
		errMsg    string
	}{
		{
			"regular url",
			"https://video-packager.example.com",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"extra slashes on endpoint",
			"https://video-packager.example.com////",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"rendition path",
			"https://video-packager.example.com",
			"/video/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"https://video-packager.example.com/video/t/2017/07/11/73302_1_12trumpjr-emails_wg_240p/",
			"",
		},
		{
			"invalid rendition url - not /video/ prefix",
			"https://video-packager.example.com",
			"https://vp.nyt.com/videos/2017/07/11/73302_1_12trumpjr-emails_wg_240p.mp4",
			"",
			"invalid rendition: path doesn't start with /video/",
		},
		{
			"invalid rendition url - not mp4",
			"https://video-packager.example.com",
			"https://vp.nyt.com/video/2017/07/11/73302_1_12trumpjr-emails_wg_480p.webm",
			"",
			"invalid rendition: not an mp4 file",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := Generator{VideoPackagerEndpoint: test.endpoint}
			prefix, err := generator.videoPackagerPrefix(test.rendition)
			if err == nil {
				err = errors.New("")
			}
			if err.Error() != test.errMsg {
				t.Errorf("wrong error\nwant %q\ngot  %q", test.errMsg, err.Error())
			}
			if prefix != test.expected {
				t.Errorf("wrong prefix\nwant %q\ngot  %q", test.expected, prefix)
			}
		})
	}
}
