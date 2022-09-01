package main

import (
	"encoding/json"
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
	"golang.design/x/clipboard"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"viedo2m3u8/model"
	"viedo2m3u8/theme"
	"viedo2m3u8/util"
)

var (
	videoExts = map[string]bool{
		".mp4": true,
		".avi": true,
	}
	settingFile = ".caisin.setting"
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
func init() {
	clipboard.Init()
}

func main() {

	a := app.New()

	setting := &model.Setting{Rate: "60"}

	str, err := files.ReadStr(settingFile)
	if err == nil {
		json.Unmarshal([]byte(str), setting)
	}
	caisinTheme := &theme.CaisinTheme{}
	a.Settings().SetTheme(caisinTheme)
	icon := caisinTheme.Icon("logo")
	//a.SetIcon(icon)
	w := a.NewWindow("视频压缩")
	w.SetIcon(icon)

	rateEntry := widget.NewEntry()
	rateEntry.SetText(setting.Rate)
	rateEntry.PlaceHolder = "帧率"

	rateEntry.OnChanged = func(s string) {
		setting.Rate = s
	}
	rateBox := container.NewHBox(widget.NewLabel("帧率:"), rateEntry)

	srcPathLabel := widget.NewLabel(setting.InPath)
	selBtn := widget.NewButton("选择", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if dir == nil {
				dialog.ShowInformation("目录", "取消选择", w)
			} else {
				setting.InPath = dir.Path()
				srcPathLabel.SetText(setting.InPath)
			}
		}, w)
	})
	srcBox := container.NewHBox(widget.NewLabel("视频目录:"), srcPathLabel, selBtn)

	url := "https://ffmpeg.org/download.html"
	info := widget.NewLabel(fmt.Sprintf("使用前请先安装ffmpeg\n下载地址:%s", url))
	infoBox := container.NewHBox(info, widget.NewButton("去下载", func() {
		util.Open(url)
	}), widget.NewButton("复制", func() {
		clipboard.Write(clipboard.FmtText, []byte(url))
		dialog.ShowInformation("成功", "复制成功", w)
	}))

	ffmPathLabel := widget.NewLabel(setting.FfmpegPath)
	ffmpegBtn := widget.NewButton("选择", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if dir == nil {
				dialog.ShowInformation("目录", "取消选择", w)
			}
			setting.FfmpegPath = dir.Path()
			ffmPathLabel.SetText(setting.FfmpegPath)
		}, w)
	})
	ffmBox := container.NewHBox(widget.NewLabel("ffmpeg目录:"), ffmPathLabel, ffmpegBtn)

	outPathLabel := widget.NewLabel(setting.OutPath)
	selOutBtn := widget.NewButton("选择", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if dir == nil {
				dialog.ShowInformation("目录", "取消选择", w)
			} else {
				setting.OutPath = dir.Path()
				if setting.OutPath == setting.InPath {
					setting.OutPath = ""
					dialog.ShowError(errors.New("输入目录和输出目录不能相同"), w)
				}
				outPathLabel.SetText(setting.OutPath)

			}
		}, w)
	})
	outBox := container.NewHBox(widget.NewLabel("输出目录:"), outPathLabel, selOutBtn)

	pro := widget.NewLabel(fmt.Sprintf("转化进度:%d/%d", 0, 0))

	saveSetBtn := widget.NewButton("保存配置", func() {
		setting.Rate = rateEntry.Text
		saveSetting(setting)
	},
	)
	hello := widget.NewCard("压缩设置", "视频目录和输出地址设置",
		container.NewVBox(ffmBox,
			rateBox,
			srcBox,
			outBox,
			pro,
			saveSetBtn),
	)

	hello.Resize(fyne.NewSize(750, 400))
	transBtn := widget.NewButton("转换", func() {})
	transBtn.OnTapped = func(transBtn *widget.Button) func() {
		return func() {
			trans(pro, transBtn, w, setting)
		}
	}(transBtn)
	w.SetContent(container.NewVBox(
		infoBox,
		hello,
		transBtn,
	))
	w.Resize(fyne.NewSize(750, 500))
	w.ShowAndRun()
}
func trans(label *widget.Label, btn *widget.Button, w fyne.Window, setting *model.Setting) {
	if strutil.IsBlank(setting.OutPath) {
		dialog.ShowError(errors.New("输出目录不能为空"), w)
		return
	}
	btn.Disable()
	defer btn.Enable()
	if !files.PathExist(setting.InPath) {
		dialog.ShowError(errors.New("输入目录不存在"), w)
		return
	}
	entries, err := os.ReadDir(setting.InPath)
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

	err = files.IsNotExistMkDir(setting.OutPath)
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
			ret := util.CompressVideo(setting, path.Join(setting.InPath, mp4), path.Join(setting.OutPath, mp4))
			succ.Add(1)
			fmt.Println(ret)
		})
	}

	wg.Wait()

	dialog.ShowInformation("成功", fmt.Sprintf("转化成功:%s,失败:%s", succ, fail), w)
}

func saveSetting(set *model.Setting) {
	marshal, err := json.Marshal(set)
	if err == nil {
		file, err := files.OpenOrCreateFile(settingFile)
		if err == nil {
			defer file.Close()
			file.Write(marshal)
		}
	}
}
