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

type NotionTools struct {
	Notion notion.NotionConfig
	Jira   jira.JiraConfig
}

var date time.Time
var tool NotionTools

func init() {
	var cp = flag.String("file", "", "Notion configuration file")
	var dp = flag.String("date", "", "Reporting date")
	flag.Parse()

	if *cp == "" {
		fmt.Println("usage: notion-tools -file='CONFIGURATION' [-date='YYYY-MM-DD']")
		os.Exit(1)
	}

	if *dp == "" {
		*dp = time.Now().Format(utils.YYYYMMDD)
	}

	d, err := time.Parse(utils.YYYYMMDD, *dp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	date = d

	_, err = toml.DecodeFile(*cp, &tool)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func main() {
	for _, ingredient := range tool.Jira.Ingredients {
		d, count, err := ingredient.Get(tool.Jira, date)
		if err != nil {
			fmt.Printf("error with ingredient '%v': %v\n", ingredient.Name, err)
			continue
		}
		tool.Notion.Report(ingredient.NotionID, d, count)
	}
}
