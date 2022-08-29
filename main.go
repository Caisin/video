package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/Caisin/caisin-go/utils/files"
	"gitee.com/Caisin/caisin-go/utils/strutil"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/atomic"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"viedo2m3u8/theme"
	"viedo2m3u8/util"
)

var (
	videoExts = map[string]bool{
		".mp4": true,
		".avi": true,
	}
)

/*
	func init() {
		sysType := runtime.GOOS
		switch sysType {
		case "windows":
		case "macos":
			util.RunCommand(".", "brew", "update", "&&", "brew", "upgrade", "ffmpeg")
		case "linux":

		}

}
*/
func main() {

	a := app.New()

	caisinTheme := &theme.CaisinTheme{}
	a.Settings().SetTheme(caisinTheme)
	icon := caisinTheme.Icon("logo")
	//a.SetIcon(icon)
	w := a.NewWindow("视频压缩")
	w.SetIcon(icon)

	srcPath := ""
	srcPathLabel := widget.NewLabel(srcPath)
	selBtn := widget.NewButton("选择", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if dir == nil {
				dialog.ShowInformation("目录", "取消选择", w)
			} else {
				srcPath = dir.Path()
				srcPathLabel.SetText(srcPath)
			}
		}, w)
	})
	srcBox := container.NewHBox(widget.NewLabel("视频目录:"), srcPathLabel, selBtn)

	info := widget.NewLabel("使用前请先安装ffmpeg\n下载地址:https://ffmpeg.org/download.html")
	outPath := ""
	outPathLabel := widget.NewLabel(outPath)
	selOutBtn := widget.NewButton("选择", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if dir == nil {
				dialog.ShowInformation("目录", "取消选择", w)
			} else {
				outPath = dir.Path()
				if outPath == srcPath {
					outPath = ""
					dialog.ShowError(errors.New("输入目录和输出目录不能相同"), w)
				}
				outPathLabel.SetText(outPath)

			}
		}, w)
	})
	outBox := container.NewHBox(widget.NewLabel("输出目录:"), outPathLabel, selOutBtn)

	pro := widget.NewLabel(fmt.Sprintf("转化进度:%d/%d", 0, 0))

	hello := widget.NewCard("压缩设置", "视频目录和输出地址设置", container.NewVBox(srcBox, outBox, pro))

	hello.Resize(fyne.NewSize(750, 400))
	transBtn := widget.NewButton("转换", func() {})
	transBtn.OnTapped = func(transBtn *widget.Button) func() {
		return func() {
			trans(pro, transBtn, w, srcPath, outPath)
		}
	}(transBtn)
	w.SetContent(container.NewVBox(
		info,
		hello,
		transBtn,
	))
	w.Resize(fyne.NewSize(750, 500))
	w.ShowAndRun()
}
func trans(label *widget.Label, btn *widget.Button, w fyne.Window, src, out string) {
	if strutil.IsBlank(out) {
		dialog.ShowError(errors.New("输出目录不能为空"), w)
		return
	}
	btn.Disable()
	defer btn.Enable()
	if !files.PathExist(src) {
		dialog.ShowError(errors.New("输入目录不存在"), w)
		return
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	mp4s := make([]string, 0)
	for _, file := range entries {
		name := file.Name()
		ext := path.Ext(name)
		if videoExts[strings.ToLower(ext)] {
			mp4s = append(mp4s, name)
		}
	}
	if len(mp4s) == 0 {
		dialog.ShowError(errors.New("输入目录无视频文件"), w)
		return
	}

	err = files.IsNotExistMkDir(out)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	wg := sync.WaitGroup{}
	p, _ := ants.NewPool(1, ants.WithPreAlloc(true))
	defer p.Release()
	succ, fail, done := atomic.NewUint32(0), atomic.NewUint32(0), atomic.NewUint32(0)
	total := len(mp4s)
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			label.SetText(fmt.Sprintf("转化进度:%s/%d", done, total))

		}
	}()
	for i := range mp4s {
		wg.Add(1)
		mp4 := mp4s[i]
		p.Submit(func() {
			defer wg.Done()
			defer done.Add(1)
			ret := util.CompressVideo(path.Join(src, mp4), path.Join(out, mp4))
			succ.Add(1)
			fmt.Println(ret)
		})
	}

	wg.Wait()

	dialog.ShowInformation("成功", fmt.Sprintf("转化成功:%s,失败:%s", succ, fail), w)
}
