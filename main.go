package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	_path := flag.String("path", "", "The `DIRPATH` to clean up. Only sub-items are inspected.")
	_before := flag.Int("sub", 30, "No. of `DAYS` ago. All items modified before that are deleted.")
	_dryrun := flag.Bool("dryrun", false, "Perform a dry run. Cleanup is not perfomed.")
	_except := flag.String("except", "", "A comma-separated list of `REGEXP` for cleanup exclusion.")
	flag.Parse()

	if *_path == "" {
		panic("No -path provided.")
	}

	files, err := ioutil.ReadDir(*_path)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	before := now.AddDate(0, 0, *_before*-1)
	log.Println("All items modified before", before.Format(time.UnixDate), "will be deleted.")

	for _, f := range files {
		p := *_path + "\\" + f.Name()
		matches := strings.Split(*_except, ",")
		ping := false
		if len(*_except) > 0 {
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
					if *_dryrun {
						log.Println(p, "will be deleted.")
					} else {
						log.Println("Deleting", p, "...")
						if mt.IsDir() {
							c := exec.Command("cmd", "/C", "rmdir", "/s", "/q", p)
							if err := c.Run(); err != nil {
								log.Println(err)
							}
						}
					}
				}
			}
		} else {
			log.Println(p, "matched. Skip.")
		}
	}
}
