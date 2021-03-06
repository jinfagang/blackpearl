package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	//"./picbed"
	//"github.com/atotto/clipboard"
	"github.com/urfave/cli/v2"
	"gitlab.com/jinfagang/colorgo"
	"gopkg.in/yaml.v2"

	"blackpearl/src/components"
	"blackpearl/src/models"

	w "blackpearl/src/widgets"
	ui "github.com/gizak/termui/v3"
)

const (
	appName                   = "blackpearl"
	version                   = "0.0.1"
	graphHorizontalScaleDelta = 3
)

var (
	stderrLogger = log.New(os.Stderr, "", 0)

	graphHorizontalScale = 7
	helpVisible          = false

	updateInterval = time.Second
	minimalMode    = false
	averageLoad    = false
	percpuLoad     = false
	// tempScale      = w.Celcius
	battery   = false
	statusbar = false
	// netInterface   = w.NET_INTERFACE_ALL

	// cpu  *w.CpuWidget
	batt *w.BatteryWidget
	// mem  *w.MemWidget
	// proc *w.ProcWidget
	// net  *w.NetWidget
	disk *w.DiskWidget
	// temp *w.TempWidget
	help *w.HelpMenu
	grid *ui.Grid
	bar  *w.StatusBar
	todo *w.TodoWidget
)

func welcome() {
	fmt.Print(cg.BoldStart)
	cg.Foreground(cg.Blue, true)
	fmt.Println("The Black Pearl with many useful tools!")
	fmt.Print(cg.BoldEnd)

	cg.Foreground(cg.Red, true)
	fmt.Println("author: Lucas Jin jinfagang19@gmail.com")
	cg.ResetColor()
	fmt.Println("http://github.com/jinfagang/blackpearl")
	fmt.Print(cg.BoldStart)
	cg.Foreground(cg.Green, true)
	fmt.Println(`     ____  __           __      ____                  __
    / __ )/ /___ ______/ /__   / __ \___  ____ ______/ /
   / __  / / __ ~/ ___/ //_/  / /_/ / _ \/ __ ~/ ___/ / 
  / /_/ / / /_/ / /__/ ,<    / ____/  __/ /_/ / /  / /  
 /_____/_/\__,_/\___/_/|_|  /_/    \___/\__,_/_/  /_/   
														`)
	fmt.Print(cg.BoldEnd)
}

func main() {
	welcome()

	// start reading config file
	var configYamlF = filepath.Join(userHomeDir(), ".conf/blackpearl.yaml")
	f, err := os.Open(configYamlF)
	var cfg models.Config
	if err != nil {
		fmt.Printf("no local config file found %s\n", configYamlF)
	} else {
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&cfg)
		if err != nil {
			fmt.Println("decode config file error, wrong attributes?")
		}
	}
	// some constant
	var apiExtBlackPearl = "http://loliloli.pro:9000/api/v2/ext_blackpearl"

	var blogTitle string
	var blogUseDate bool
	var uploadImgPath string
	var uploadClipboard bool
	var pushContent string
	var pushBindUserAcc string
	var pushUseMemory bool

	app := &cli.App{
		Name:  "bp",
		Usage: "help you finish some dirty jobs. in pure go.",
		Action: func(c *cli.Context) error {
			fmt.Println("you have not set command, type -h for details.")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:     "blog",                       //命令名称
				Aliases:  []string{"b"},                // 命令的别名列表
				Usage:    "blackpearl blog blog_title", // 命令描述
				Category: "Writting",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "title",    // 配置名称
						Value:       "no_title", // 缺省配置值
						Aliases:     []string{"t"},
						Usage:       "Blog title", // 配置描述
						Destination: &blogTitle,   // 保存配置值
					},
					&cli.BoolFlag{
						Name:        "date", // 配置名称
						Aliases:     []string{"d"},
						Usage:       "Set using date as blog title prefix.", // 配置描述
						Destination: &blogUseDate,                           // 保存配置值
					},
				},
				Action: func(c *cli.Context) error { // 函数入口
					lang := components.DetectLang(blogTitle)
					if blogUseDate {
						fmt.Print(cg.BoldStart)
						cg.Foreground(cg.Green, true)
						fmt.Println("user date ON.")
						fmt.Print(cg.BoldEnd)
						var datePrefix = time.Now().Format("2006_01_02_03_")
						blogTitle = datePrefix + blogTitle
					}
					blog := components.Blog{
						Title:         blogTitle,
						Author:        components.GetCurrentUserName(),
						CreateTime:    components.GetNowTimeString(),
						Category:      components.IntelligentCategorize(blogTitle, lang),
						Summary:       components.GenerateSummary(blogTitle, lang),
						SavePath:      "./",
						CopyrightInfo: components.GetCopyrightInfo(lang),
					}

					components.WriteBlog(blog)
					fmt.Print(cg.BoldStart)
					cg.Foreground(cg.Yellow, true)
					fmt.Println("write blog templates success!")
					fmt.Print(cg.BoldEnd)
					return nil
				},
			},
			{
				Name:     "upload",
				Aliases:  []string{"u"},
				Usage:    "upload picture to picbed",
				Category: "Writting",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path", // 配置名称
						Value:       "",     // 缺省配置值
						Aliases:     []string{"p"},
						Usage:       "image path",   // 配置描述
						Destination: &uploadImgPath, // 保存配置值
					},
					&cli.BoolFlag{
						Name:        "clipboard", // 配置名称
						Aliases:     []string{"c"},
						Usage:       "Uploading image copies from clipboard.", // 配置描述
						Destination: &uploadClipboard,                         // 保存配置值
					},
				},
				Action: func(c *cli.Context) error {
					// upload picture
					if uploadClipboard {
						fmt.Println("uploading images data from clipboard (not supported for now).")
					} else {
						if Exists(uploadImgPath) {
							// upload this image using picbed
							fmt.Printf("uploading from: %s\n", uploadImgPath)
							// imgBytes, err := ioutil.ReadFile(uploadImgPath)
							// if err != nil {
							// 	fmt.Println(err)
							// }
							// imgParam := picbed.ImageParam{
							// 	Name:    filepath.Base(uploadImgPath),
							// 	Type:    "jpg",
							// 	Content: &imgBytes,
							// }
							// // using Ali as default
							// client := picbed.Ali{FileLimit: nil, MaxSize: 5024}
							// res, _ := client.Upload(&imgParam)
							// fmt.Println(res.Url)
							// clipboard.WriteAll(res.Url)
							fmt.Println("result url has been copied to your clipboard.")
						} else {
							fmt.Printf("%s does not exist!\n", uploadImgPath)
						}
					}
					return nil
				},
			},
			{
				Name:     "code",
				Usage:    "init a code project",
				Category: "Develop",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:     "push",
				Aliases:  []string{"p"},
				Usage:    "push text or image to your Phone through Uranus",
				Category: "UranusMessage",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "content", // 配置名称
						Value:       "",        // 缺省配置值
						Aliases:     []string{"c"},
						Usage:       "string or image/file path", // 配置描述
						Destination: &pushContent,                // 保存配置值
					},
					&cli.StringFlag{
						Name:        "bind", // 配置名称
						Value:       "",     // 缺省配置值
						Aliases:     []string{"b"},
						Usage:       "UserAcc you bind for receive messages.", // 配置描述
						Destination: &pushBindUserAcc,                         // 保存配置值
					},
					&cli.BoolFlag{
						Name:        "memory, m",                     // 配置名称
						Usage:       "Pushing image through memory.", // 配置描述
						Destination: &pushUseMemory,                  // 保存配置值
					},
				},
				Action: func(c *cli.Context) error {
					if pushBindUserAcc != "" {
						// save bind
						fmt.Printf("[uranus bind] binding receive user acc: %s\n", pushBindUserAcc)
						// var configData = `
						// 	networks:
						// 	my_network:
						// 		driver: bridge
						// 		`
					}
					// upload picture
					if Exists(pushContent) {
						// TODO: only image support for now
						// upload this image using picbed
						fmt.Printf("[uranus push] image file, uploading from: %s\n", pushContent)
						// imgBytes, err := ioutil.ReadFile(pushContent)
						// if err != nil {
						// 	fmt.Println(err)
						// }
						// imgParam := picbed.ImageParam{
						// 	Name:    filepath.Base(pushContent),
						// 	Type:    "jpg",
						// 	Content: &imgBytes,
						// }
						// // using Ali as default
						// client := picbed.Ali{FileLimit: nil, MaxSize: 5024}
						// res, _ := client.Upload(&imgParam)
						// pushContent = res.Url
					} else if pushUseMemory {
						// upload image to url from memory
					} else {
						// normal text
						fmt.Println("[uranus push] pushing normal text content to uranus.")
					}
					var targetAcc = pushBindUserAcc
					fmt.Printf("message pushing to uranus, target: %s, content: %s\n", targetAcc, pushContent)
					// call api/v2/ext_blackpearl here
					// target_user_acc && content
					// do a post with url
					resp, err := http.PostForm(apiExtBlackPearl, url.Values{"target_acc": {targetAcc}, "content": {pushContent}})
					if err != nil {
						fmt.Printf("Got error from server: %s\n", err.Error())
					} else {
						defer resp.Body.Close()
						body, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							fmt.Println(err.Error())
						} else {
							var data map[string]interface{}
							err := json.Unmarshal(body, &data)
							if err != nil {
								panic(err)
							}
							fmt.Println(data)
						}
					}
					return nil
				},
			},
		},
	}

	if len(os.Args) > 1 {
		err = app.Run(os.Args)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("enter panel...")
		if err := ui.Init(); err != nil {
			stderrLogger.Fatalf("failed to initialize termui: %v", err)
		}
		defer ui.Close()

		initWidgets()
		setupGrid()

		termWidth, termHeight := ui.TerminalDimensions()
		if statusbar {
			grid.SetRect(0, 0, termWidth, termHeight-1)
		} else {
			grid.SetRect(0, 0, termWidth, termHeight)
		}
		help.Resize(termWidth, termHeight)

		ui.Render(grid)
		if statusbar {
			bar.SetRect(0, termHeight-1, termWidth, termHeight)
			ui.Render(bar)
		}

		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		defer ui.Close()

		eventLoop()

	}

}

func setupGrid() {
	grid = ui.NewGrid()

	var row1 ui.GridItem
	row1 = ui.NewRow(1.0/3,
		ui.NewCol(2.0/3, todo),
		ui.NewCol(1.0/3, batt),
	)
	grid.Set(
		row1,
		ui.NewRow(1.0/3,
			ui.NewCol(1.0/3,
				ui.NewRow(1.0/2, disk),
				ui.NewRow(1.0/2, todo),
			),
			ui.NewCol(2.0/3, batt),
		),
	)
}

func initWidgets() {
	batt = w.NewBatteryWidget(graphHorizontalScale)
	bar = w.NewStatusBar()
	todo = w.NewTodoWidget()
	// proc = w.NewProcWidget()
	disk = w.NewDiskWidget()

	help = w.NewHelpMenu()

}

func eventLoop() {
	drawTicker := time.NewTicker(updateInterval).C

	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	uiEvents := ui.PollEvents()

	previousKey := ""

	for {
		select {
		case <-sigTerm:
			return
		case <-drawTicker:
			if !helpVisible {
				ui.Render(grid)
				if statusbar {
					ui.Render(bar)
				}
			}
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "?":
				helpVisible = !helpVisible
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				termWidth, termHeight := payload.Width, payload.Height
				if statusbar {
					grid.SetRect(0, 0, termWidth, termHeight-1)
					bar.SetRect(0, termHeight-1, termWidth, termHeight)
				} else {
					grid.SetRect(0, 0, payload.Width, payload.Height)
				}
				help.Resize(payload.Width, payload.Height)
				ui.Clear()
			}

			if helpVisible {
				switch e.ID {
				case "?":
					ui.Clear()
					ui.Render(help)
				case "<Escape>":
					helpVisible = false
					ui.Render(grid)
				case "<Resize>":
					ui.Render(help)
				}
			} else {
				switch e.ID {
				case "?":
					ui.Render(grid)
				case "h":
					graphHorizontalScale += graphHorizontalScaleDelta
					// cpu.HorizontalScale = graphHorizontalScale
					// mem.HorizontalScale = graphHorizontalScale
					// ui.Render(cpu, mem)
				case "l":
					if graphHorizontalScale > graphHorizontalScaleDelta {
						graphHorizontalScale -= graphHorizontalScaleDelta
						// cpu.HorizontalScale = graphHorizontalScale
						// mem.HorizontalScale = graphHorizontalScale
						// ui.Render(cpu, mem)
					}
				case "<Resize>":
					ui.Render(grid)
					if statusbar {
						ui.Render(bar)
					}
				case "<MouseLeft>":
					payload := e.Payload.(ui.Mouse)
					todo.HandleClick(payload.X, payload.Y)
					ui.Render(todo)
				case "k", "<Up>", "<MouseWheelUp>":
					todo.ScrollUp()
					ui.Render(todo)
				case "j", "<Down>", "<MouseWheelDown>":
					todo.ScrollDown()
					ui.Render(todo)
				case "<Home>":
					todo.ScrollTop()
					ui.Render(todo)
				case "g":
					if previousKey == "g" {
						todo.ScrollTop()
						ui.Render(todo)
					}
				case "G", "<End>":
					todo.ScrollBottom()
					ui.Render(todo)
				case "<C-d>":
					todo.ScrollHalfPageDown()
					ui.Render(todo)
				case "<C-u>":
					todo.ScrollHalfPageUp()
					ui.Render(todo)
				case "<C-f>":
					todo.ScrollPageDown()
					ui.Render(todo)
				case "<C-b>":
					todo.ScrollPageUp()
					ui.Render(todo)
				case "d":
					if previousKey == "d" {
						// proc.KillProc()
					}
				case "<Tab>":
					// proc.ToggleShowingGroupedProcs()
					ui.Render(todo)
				case "m", "c", "p":
					// proc.ChangeProcSortMethod(w.ProcSortMethod(e.ID))
					ui.Render(todo)
				}

				if previousKey == e.ID {
					previousKey = ""
				} else {
					previousKey = e.ID
				}
			}

		}
	}
}
