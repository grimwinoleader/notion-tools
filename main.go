package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/notion-tools/jira"
	"github.com/notion-tools/notion"
	"github.com/notion-tools/utils"
)

var date time.Time
var cfg notion.NotionConfig

func init() {
	var cfgFilePtr = flag.String("file", "", "Notion configuration file")
	var datePtr = flag.String("date", "", "Reporting date")
	flag.Parse()

	if *cfgFilePtr == "" {
		fmt.Println("A Notion configuration file must be specified.\n")
		fmt.Println("usage: notion-tools -file='CONFIGURATION' [-date='YYYY-MM-DD']")
		os.Exit(1)
	}

	if *datePtr == "" {
		*datePtr = time.Now().Format(utils.YYYYMMDD)
	}

	d, err := time.Parse(utils.YYYYMMDD, *datePtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	date = d

	_, err = toml.DecodeFile(*cfgFilePtr, &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	fmt.Println("Reporting period ending ", date.Format(utils.YYYYMMDD))

	for _, ingredient := range cfg.Jira.Ingredients {
		d, count := jira.Search(cfg.Jira, ingredient, date)
		notion.Report(cfg, ingredient.NotionID, d, count)
	}

}
