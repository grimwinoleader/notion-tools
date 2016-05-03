package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	YYYYMMDD = "2006-01-02"
)

type NotionConfig struct {
	Token string
	URL   string
	Jira  JiraConfig
}

type NotionResponse struct {
	Errors []string `json:"errors"`
	Status string   `json:"status"`
}

type JiraConfig struct {
	Username    string
	Password    string
	Url         string
	Ingredients []JiraIngredient
}

type JiraIngredient struct {
	Name     string
	NotionID string
	Freq     string
	JQL      string
}

type JiraSearchResponse struct {
	Total      int `json:"total"`
	MaxResults int `json:"maxResults"`
	StartAt    int `json:"startAt"`
}

func Search(cfg JiraConfig, ingredient JiraIngredient, end time.Time) (time.Time, int) {

	var u *url.URL
	u, _ = url.Parse(cfg.Url + "/rest/api/latest/search")

	start := end
	if ingredient.Freq == "w" {
		end = StartOfWeek(end)
		start = end.AddDate(0, 0, -7)
	}
	jql := fmt.Sprintf(ingredient.JQL, start.Format(YYYYMMDD), end.Format(YYYYMMDD))

	payload := fmt.Sprintf(`{ "jql":"%v", "startAt":0, "fields":["key"] }`, jql)

	var jsonStr = []byte(payload)
	r, _ := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonStr))
	r.SetBasicAuth(cfg.Username, cfg.Password)
	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var jsr JiraSearchResponse
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &jsr)
	if err != nil {
		fmt.Println("unexpected response from jira. ", err)
		os.Exit(100)
	}
	return end, jsr.Total
}

func Report(cfg NotionConfig, id string, date time.Time, val int) {

	payload := fmt.Sprintf(`{ "ingredient_id":"%v", "date":"%v", "value":%v }`, id, date.Format(YYYYMMDD), val)
	fmt.Println(payload)

	var jsonStr = []byte(payload)
	r, _ := http.NewRequest("POST", cfg.URL, bytes.NewBuffer(jsonStr))
	r.Header.Set("Authorization", cfg.Token)
	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		fmt.Println("unexpected status from notion", resp.Status)
		os.Exit(100)
	}

	var nr NotionResponse
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &nr)
	if err != nil {
		fmt.Println("unexpected body from notion. ", err)
		os.Exit(100)
	}
}

func StartOfWeek(date time.Time) time.Time {
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

var date time.Time
var cfg NotionConfig

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
		*datePtr = time.Now().Format(YYYYMMDD)
	}

	d, err := time.Parse(YYYYMMDD, *datePtr)
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
	fmt.Println("Reporting period ending ", date.Format(YYYYMMDD))

	for _, ingredient := range cfg.Jira.Ingredients {
		d, count := Search(cfg.Jira, ingredient, date)
		Report(cfg, ingredient.NotionID, d, count)
	}

}
