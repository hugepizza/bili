package pkgs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MergeAV 合并音视频
func MergeAV(f1, f2 string) {
	outName := f1[0:strings.LastIndex(f1, ".")]
	cmd := exec.Command("bash", "-c", fmt.Sprintf("ffmpeg -i %s -i %s -c copy %s_merged.mp4 -y", f1, f2, outName))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Start()
	cmd.Wait()
}

// Check 检查是否安装ffmpeg
func Check() error {
	cmd := exec.Command("bash", "-c", "ffmpegc")
	w := bytes.NewBuffer(nil)
	cmd.Stderr = w
	cmd.Run()
	if !strings.Contains(string(w.Bytes()), "usage") {
		return errors.New("没有安装ffmpeg")
	}
	return nil
}
