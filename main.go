package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"scan-port/ip"
	"strconv"
	"strings"
)

var basePath string
var configPath string

var portListData = []string{}

// 初始化软件配置
func initSoftConfig(config *Config) {
	goos := runtime.GOOS
	if goos == "windows" {
		dir, _ := os.Getwd()
		basePath = dir + "/data"
		os.Mkdir(basePath, os.ModePerm)
	} else {
		user, _ := user.Current()
		basePath = user.HomeDir + "/soft/" + config.DirectoryName
		os.MkdirAll(basePath, os.ModePerm)
	}

	configPath = basePath + "/config.ini"

	_, err := os.Stat(configPath)
	if err != nil {
		os.Create(configPath)
	}

	cfg, _ := ini.Load(configPath)

	fontPath := cfg.Section("").Key("font_path").String()
	setChineseFont(fontPath)
}

// 设置中文字体
func setChineseFont(path string) {
	if path == "" {
		goos := runtime.GOOS
		if goos == "windows" {
			os.Setenv("FYNE_FONT", "c:/windows/fonts/Msyh.ttc")
		} else {
			os.Setenv("FYNE_FONT", "/System/Library/Fonts/STHeiti Light.ttc")
		}
	} else {
		os.Setenv("FYNE_FONT", path)
	}
}

func main() {
	config, updateInfo := InitConfig()

	initSoftConfig(config)

	ipv4 := ip.GetIpv4()
	println(ipv4)

	// ip.FindProcess(3306)

	a := app.New()
	w := a.NewWindow(config.Title + " v" + config.Version)
	w.SetFixedSize(true)
	w.Resize(fyne.Size{Width: 500, Height: 400})
	w.SetIcon(config.Icon)

	cfg, _ := ini.Load(configPath)

	// 扫描端口标签
	ipInput := widget.NewEntry()
	startPortInput := widget.NewEntry()
	endPortInput := widget.NewEntry()

	scanTipLabel := widget.NewLabel("开始扫描吧!!!")
	portList := widget.NewList(
		func() int {
			return len(portListData)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.InfoIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(portListData[id])
		},
	)
	portList.OnSelected = func(id widget.ListItemID) {
		labelText := portListData[id]
		port, _ := strconv.Atoi(strings.Split(labelText, ":")[1])
		pidArr := ip.FindProcess(port)

		if len(pidArr) > 0 {
			dialog.ShowConfirm("提示", "确定关闭对应进程吗 ?", func(b bool) {
				if b {
					for _, pid := range pidArr {
						ip.KillProcess(pid)
					}
					dialog.ShowInformation("info", "已关闭 !!!", w)
				}
			}, w)
		} else {
			dialog.ShowInformation("info", "没有找到对应需要关闭的进程 !!!", w)
		}

		portList.Unselect(id)
	}
	listBox := container.NewVBox(
		scanTipLabel,
		container.NewGridWrap(fyne.Size{Width: 200, Height: 300}, portList),
	)

	scanPortButton := widget.NewButton("开始扫描", func() {
		if ipInput.Text == "" {
			dialog.ShowInformation("info", "请输入要扫描的IP", w)
			return
		}
		startPort, startPortInputErr := strconv.Atoi(startPortInput.Text)
		if startPortInputErr != nil {
			dialog.ShowInformation("info", "请正确输入要扫描的开始端口号", w)
			return
		}
		if startPort <= 0 {
			dialog.ShowInformation("info", "请正确输入要扫描的开始端口号, 必须大等于0", w)
			return
		}
		endPort, endPortInputErr := strconv.Atoi(endPortInput.Text)
		if endPortInputErr != nil {
			dialog.ShowInformation("info", "请正确输入要扫描的结束端口号", w)
			return
		}
		if endPort <= 0 {
			dialog.ShowInformation("info", "请正确输入要扫描的结束端口号, 必须大等于0", w)
			return
		}
		portGap := endPort - startPort
		if portGap < 0 {
			dialog.ShowInformation("info", "结束端口号不能大于开始端口号", w)
			return
		}

		scanTipLabel.SetText("正在扫描中...")
		portListData = []string{}

		portInfos := make(chan ip.PortInfo, portGap+1)
		ip.ScanPort(portInfos, ipInput.Text, startPort, endPort)
		notAvailablePort := 0
		for i := 0; i < (portGap + 1); i++ {
			portInfo := <-portInfos
			if !portInfo.Available {
				notAvailablePort += 1
				portListData = append(portListData, portInfo.Ip+":"+strconv.Itoa(portInfo.Port))
			}
		}

		if notAvailablePort > 0 {
			scanTipLabel.SetText("扫描完毕, 已发现 " + strconv.Itoa(notAvailablePort) + " 个占用端口...")
		} else {
			scanTipLabel.SetText("扫描完毕, 未发现被占用的端口...")
		}
	})

	ipInput.SetText(ipv4)

	updateVersionButton := widget.NewButton("新版本更新", func() {
		dialog.ShowInformation("新版本 v"+updateInfo.Version, updateInfo.Description, w)
	})
	if updateInfo.Version == config.Version {
		updateVersionButton.Hidden = true
	} else {
		updateVersionButton.Hidden = false
	}

	ipInputBox := container.NewVBox(
		widget.NewLabel("ip地址"),
		container.NewGridWrap(fyne.Size{Height: 35, Width: 150}, ipInput),
		widget.NewLabel("开始端口"),
		container.NewGridWrap(fyne.Size{Height: 35, Width: 100}, startPortInput),
		widget.NewLabel("结束端口"),
		container.NewGridWrap(fyne.Size{Height: 35, Width: 100}, endPortInput),
		scanPortButton,
		updateVersionButton,
	)

	scanPortBox := container.NewHBox(
		ipInputBox,
		container.NewGridWrap(fyne.Size{Height: 35, Width: 50}),
		listBox,
	)

	// 软件信息标签
	fontPathText := widget.NewMultiLineEntry()
	fontPathText.SetText(cfg.Section("").Key("font_path").String())
	fontForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "字体文件路径", Widget: fontPathText},
		},
		OnSubmit: func() {
			cfg.Section("").Key("font_path").SetValue(fontPathText.Text)
			cfg.SaveTo(configPath)
			dialog.ShowInformation("info", "保存成功", w)
		},
	}

	giteeUrl, _ := url.Parse("https://gitee.com/orange-resource/scan-port")
	githubUrl, _ := url.Parse("https://github.com/orange-resource/scan-port")
	downloadUrl, _ := url.Parse("https://gitee.com/orange-resource/scan-port/releases")
	softInfoVBox := container.NewVBox(
		fontForm,
		widget.NewLabel("软件作者: 橘子"),
		widget.NewLabel("联系QQ: 1067357662"),
		container.NewHBox(
			widget.NewHyperlink("Windows | Mac 版本下载地址", downloadUrl),
		),
		container.NewHBox(
			widget.NewHyperlink("码云开源地址", giteeUrl),
			widget.NewHyperlink("Github开源地址", githubUrl),
		),
	)

	// 总布局
	tabs := container.NewAppTabs(
		container.NewTabItem("扫描端口", scanPortBox),
		container.NewTabItem("软件信息", softInfoVBox),
	)
	w.SetContent(tabs)

	w.ShowAndRun()
}
