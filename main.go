// +build windows
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/urfave/cli"
	"gopkg.in/natefinch/lumberjack.v2"
)

const internalVersion = "1.0"

var (
	rlf *log.Logger
)

func initRotatingLog(out io.Writer) {
	rlf = log.New(out, "DIRCLEAN: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func dtrace(v ...interface{}) {
	log.Println(v...)
	rlf.Println(v...)
}

func main() {
	// Initialize rotating logs
	wd := os.TempDir()
	wd, _ = os.Getwd()
	initRotatingLog(&lumberjack.Logger{
		Dir:        wd,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     30,
	})

	app := cli.NewApp()
	app.Name = "dirclean"
	app.Usage = "Directory cleanup tool for Windows."
	app.Version = internalVersion
	app.Copyright = "(c) 2016 Chew Esmero."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path",
			Value: "",
			Usage: "the [`path`] to cleanup",
		},
		cli.IntFlag{
			Name:  "sub",
			Value: 30,
			Usage: "all items modified before [`days`] are deleted",
		},
		cli.BoolFlag{
			Name:  "dryrun",
			Usage: "perform a dry run (cleanup not performed)",
		},
		cli.StringFlag{
			Name:  "except",
			Value: "",
			Usage: "comma-separated list of [`regexp`] for exclusion",
		},
	}
	app.Action = func(c *cli.Context) error {
		if !c.GlobalIsSet("path") {
			log.Println("Flag 'path' not set.")
			return cli.NewExitError("Flag 'path' not set.", -1)
		}

		dryrun := false
		if c.GlobalIsSet("dryrun") {
			dryrun = c.GlobalBool("dryrun")
		}

		path := c.GlobalString("path")
		sub := c.GlobalInt("sub")
		except := c.GlobalString("except")
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Println(err)
			return cli.NewExitError(err.Error(), -1)
		}

		now := time.Now()
		before := now.AddDate(0, 0, sub*-1)
		dtrace("All items modified before", before.Format(time.UnixDate), "will be deleted.")

		for _, f := range files {
			p := path + "\\" + f.Name()
			matches := strings.Split(except, ",")
			ping := false
			if len(except) > 0 {
				for _, m := range matches {
					ping, _ = regexp.MatchString(m, p)
					if ping {
						ping = true
						break
					}
				}
			}

			if !ping {
				mt, err := os.Stat(p)
				if err != nil {
					log.Println(err)
				} else {
					if mt.ModTime().Before(before) {
						if dryrun {
							dtrace(p, "will be deleted.")
						} else {
							dtrace("Deleting", p, "...")
							if mt.IsDir() {
								c := exec.Command("cmd", "/C", "rmdir", "/s", "/q", p)
								if err := c.Run(); err != nil {
									dtrace(err)
								}
							}
						}
					}
				}
			} else {
				dtrace(p, "matched. Skip.")
			}
		}

		return nil
	}

	app.Run(os.Args)
}
