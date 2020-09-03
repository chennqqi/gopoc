package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"gopoc/lib"
	"gopoc/utils"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	num        int
	rate       int
	timeout    time.Duration
	proxy      string
	pocName    string
	pocDir     string
	target     string
	targetFile string
	rawFile    string
	cookie     string
	verbose    bool
	debug      bool
	forceSSL   bool

	reverse string
)

func Execute() {
	app := &cli.App{
		Name:    "go poc scanner",
		Usage:   "A poc framework written in golang",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "reverse", Aliases: []string{"e"}, Destination: &reverse, Value: "", Usage: "set reverse URI, eg: ceye://ceye.io?api=xxxxx, godnslog://godnslog.com?secret=xxxx"},
			&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}, Destination: &debug, Value: false, Usage: "log level debug"},
			&cli.BoolFlag{Name: "info", Aliases: []string{"i"}, Destination: &verbose, Value: false, Usage: "log level info"},
			&cli.StringFlag{Name: "poc-name", Aliases: []string{"p"}, Destination: &pocName, Value: "", Usage: "single poc `NAME`"},
			&cli.StringFlag{Name: "poc-dir", Aliases: []string{"P"}, Destination: &pocDir, Value: "", Usage: "load multi poc from `DIRECTORY`, eg: pocs/* or pocs/thinkphp.*"},
			&cli.StringFlag{Name: "target", Aliases: []string{"t"}, Destination: &target, Value: "", Usage: "target to scan"},
			&cli.StringFlag{Name: "targetFile", Aliases: []string{"l"}, Destination: &targetFile, Value: "", Usage: "load targets from `FILE`"},
			&cli.StringFlag{Name: "raw", Aliases: []string{"r"}, Destination: &rawFile, Value: "", Usage: "request raw `File`"},
			&cli.BoolFlag{Name: "ssl", Destination: &forceSSL, Value: false, Usage: "force usage of SSL/HTTPS for raw"},
			&cli.IntFlag{Name: "num", Value: 10, Destination: &num, Usage: "threads `NUM`"},
			&cli.IntFlag{Name: "rate", Value: 100, Destination: &rate, Usage: "scan rate, request per second"},
			&cli.DurationFlag{Name: "timeout", Destination: &timeout, Value: 10 * time.Second, Usage: "scan `TIMEOUT`"},
			&cli.StringFlag{Name: "cookie", Destination: &cookie, Value: "", Usage: "http cookie header"},
			&cli.StringFlag{Name: "proxy", Destination: &proxy, Value: "", Usage: "http proxy", DefaultText: "http://127.0.0.1:8080"},
		},
		Action: func(c *cli.Context) error {
			err := lib.InitHttpClient(num, proxy, timeout)
			if err != nil {
				return err
			}
			utils.InitLog(debug, verbose)

			if !lib.InitReverse(reverse) {
				utils.Warning("no dnslog service set")
			}
			switch {
			case target != "":
				if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
				} else {
					target = "http://" + target
				}
				req, err := http.NewRequest("GET", target, nil)
				if err != nil {
					return err
				}
				if cookie != "" {
					req.Header.Set("Cookie", cookie)
				}
				if pocName != "" {
					if poc := lib.CheckSinglePoc(req, pocName); poc != nil {
						utils.Green("%v, %s", target, poc.Name)
					}
				} else if pocDir != "" {
					lib.CheckMultiPoc(req, pocDir, num)
				}
			case targetFile != "":
				targets := utils.ReadingLines(targetFile)
				if pocName != "" {
					lib.BatchCheckSinglePoc(targets, pocName, num)
				} else if pocDir != "" {
					lib.BatchCheckMultiPoc(targets, pocDir, num, rate)
				}
			case rawFile != "":
				raw, err := ioutil.ReadFile(rawFile)
				if err != nil {
					return err
				}
				req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(raw)))
				if err != nil {
					return err
				}
				if !req.URL.IsAbs() {
					scheme := "http"
					if forceSSL {
						scheme = "https"
					}
					req.URL.Scheme = scheme
					req.URL.Host = req.Host
				}

				if pocName != "" {
					if poc := lib.CheckSinglePoc(req, pocName); poc != nil {
						utils.Green("%v, %s", target, poc.Name)
					}
				} else if pocDir != "" {
					lib.CheckMultiPoc(req, pocDir, num)
				}
			default:
				fmt.Println("Use -h for basic help")
			}
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}
}
