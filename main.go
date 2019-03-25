package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/hugepizza/bilibili/pkgs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	avNumber int
	savePath string
	merge    bool

	rootCmd = &cobra.Command{
		Use: "b-get",
		Run: run,
	}
)

func init() {
	rootCmd.PersistentFlags().IntVar(&avNumber, "av", 0, "视频av号")
	rootCmd.PersistentFlags().StringVar(&savePath, "path", "/home/wanlei/Videos", "保存路径")
	rootCmd.PersistentFlags().BoolVarP(&merge, "merge", "M", false, "是否自动合并")
}

func main() {
	rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) {
	cli := &http.Client{}
	title, urls, err := pkgs.Resolve(cli, fmt.Sprintf("https://www.bilibili.com/video/av%d", avNumber))
	if err != nil {
		logrus.Fatalf("resolve failed, %s", err)
		return
	}
	wg := sync.WaitGroup{}
	ech := make(chan error)
	go func(ch chan error) {
		select {
		case err := <-ech:
			logrus.Fatalf("download failed,%s", err)
		}
	}(ech)
	paths := []string{}
	for _, url := range urls {
		path := filepath.Join(savePath, title)

		if url.MimeType == pkgs.MimeTypeVideo {
			logrus.Info("视频下载中...")
			path += ".mp4"
		}
		if url.MimeType == pkgs.MimeTypeAudio {
			logrus.Info("音频下载中...")
			path += ".m4s"
		}
		paths = append(paths, path)
		wg.Add(1)
		go pkgs.DownloadVideo(cli, &wg, ech, url.BaseURL, avNumber, path)
	}
	wg.Wait()
	logrus.Infof("下载完成,路径:%s", savePath)
	logrus.Infof("合并中", savePath)
	if len(paths) > 1 {
		pkgs.MergeAV(paths[0], paths[1])
		logrus.Infof("合并完成,路径:%s", savePath)
	}

	return
}
