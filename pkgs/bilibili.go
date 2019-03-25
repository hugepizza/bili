package pkgs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// MimeType 媒体类型
type MimeType string

// 音视频
const (
	userAgent              = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36"
	MimeTypeAudio MimeType = "audio/mp4"
	MimeTypeVideo          = "video/mp4"

	// bilibili 视频清晰度
	p1080plus = 112
	p1080     = 80
	p720      = 64
	p480      = 32
	p360      = 16
	// bilibili 音频清晰度
	k192 = 30280
	k65  = 30216
)

type avURL struct {
	AID      int      `json:"aid"`
	ID       int      `json:"id"`
	BaseURL  string   `json:"baseUrl"`
	CodeID   int      `json:"codecid"`
	MimeType MimeType `json:"mime_type"`
}

type video struct {
	Code int `json:"code"`
	Data struct {
		AcceptQuality []int  `json:"accept_quality"`
		VideoCodecid  int    `json:"video_codecid"`
		Dash          dash   `json:"dash"`
		DURL          []durl `json:"durl"`
	} `json:"data"`
}

// m4s 格式的视频信息
type dash struct {
	Video []avURL `json:"video"`
	Audio []avURL `json:"audio"`
}

// flv 格式
type durl struct {
	URL string `json:"url"`
}

// 获取一个视频内质量最佳的音视频链接
func (v *video) getBestQuality() map[MimeType]avURL {
	if len(v.Data.Dash.Video) > 0 {
		return doM4S(v)
	}
	if len(v.Data.DURL) > 0 {
		return doFLV(v)
	}
	return nil
}
func doFLV(v *video) map[MimeType]avURL {
	if v.Data.AcceptQuality == nil || len(v.Data.AcceptQuality) < 1 {
		return nil
	}
	max := v.Data.AcceptQuality[len(v.Data.AcceptQuality)-1]
	for _, q := range v.Data.AcceptQuality {
		if q > p1080 {
			continue
		}
		if q > max {
			max = q
		}
	}
	m := make(map[MimeType]avURL)
	m[MimeTypeVideo] = avURL{
		ID:       max,
		BaseURL:  flvURL(v.Data.DURL[0].URL, max),
		MimeType: MimeTypeVideo,
	}

	return m

}

// 拼接flv最佳画质的url
func flvURL(base string, max int) string {
	logrus.Debugf("get flv url id:%d ,base:%s", max, base)
	suffix := ""
	if strings.Contains(base, ".mp4") {
		suffix = ".mp4"
	} else {
		suffix = ".flv"

	}
	return base[0:strings.Index(base, "-1-")+len("-1-")] + strconv.Itoa(max) + base[strings.Index(base, suffix):]
}

func doM4S(v *video) map[MimeType]avURL {
	if v.Data.AcceptQuality == nil || len(v.Data.AcceptQuality) < 1 {
		return nil
	}
	m := make(map[MimeType]avURL)

	if len(v.Data.Dash.Video) > 0 {
		m[MimeTypeVideo] = v.Data.Dash.Video[len(v.Data.Dash.Video)-1]
		for _, video := range v.Data.Dash.Video {
			if video.ID > p1080 {
				continue
			}
			if video.ID > m[MimeTypeVideo].ID {
				m[MimeTypeVideo] = video
			}
		}
	}
	if len(v.Data.Dash.Audio) > 0 {
		m[MimeTypeAudio] = v.Data.Dash.Audio[len(v.Data.Dash.Audio)-1]
		for _, audio := range v.Data.Dash.Audio {
			if audio.ID > m[MimeTypeAudio].ID {
				m[MimeTypeAudio] = audio
			}
		}
	}

	return m
}

// Resolve 解析网页
func Resolve(cli *http.Client, videoPageURL string) (string, map[MimeType]avURL, error) {
	req, err := http.NewRequest("GET", videoPageURL, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := cli.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return "", nil, err
	}
	videoURL := &video{}
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(s.Text(), "window.__playinfo__=") {
			rawViedo := s.Text()
			rawViedo = strings.TrimPrefix(rawViedo, "window.__playinfo__=")
			err := json.Unmarshal([]byte(rawViedo), videoURL)
			if err != nil && videoURL.Code != 0 {
				return
			}
		}
	})
	title := doc.Find(".video-title").Text()
	title = strings.Replace(title, "\\", "", -1)
	title = strings.Replace(title, "/", "", -1)
	title = strings.Replace(title, " ", "", -1)

	return title, videoURL.getBestQuality(), nil
}

// DownloadVideo 下载视频
func DownloadVideo(cli *http.Client, wg *sync.WaitGroup, ech chan error, vedioURL string, aid int, path string) error {
	defer wg.Done()
	req, err := http.NewRequest("GET", vedioURL, nil)
	if err != nil {
		ech <- err
	}
	withVideoHeader(req, aid)
	resp, err := cli.Do(req)
	if err != nil {
		ech <- err
	}
	defer resp.Body.Close()
	if resp.ContentLength < 1024 {
		return fmt.Errorf("请求返回失败,%d", resp.StatusCode)
	}
	output, err := os.Create(path)
	if err != nil {
		ech <- err
	}
	_, err = io.Copy(output, resp.Body)
	if err != nil {
		ech <- err
	}
	return nil
}

func withVideoHeader(req *http.Request, aid int) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "upos-hz-mirrorks3u.acgvideo.com")
	req.Header.Set("Origin", "https://www.bilibili.com")
	req.Header.Set("Referer", fmt.Sprintf("https://www.bilibili.com/video/av%d", aid))
	req.Header.Set("User-Agent", userAgent)
}
