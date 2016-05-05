package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/notion-tools/utils"
)

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

func (i *JiraIngredient) Get(c JiraConfig, end time.Time) (time.Time, int, error) {

	var u *url.URL
	u, _ = url.Parse(c.Url + "/rest/api/latest/search")

	start := end
	if i.Freq == "w" {
		end = utils.StartOfWeek(end)
		start = end.AddDate(0, 0, -7)
	}
	jql := fmt.Sprintf(i.JQL, start.Format(utils.YYYYMMDD), end.Format(utils.YYYYMMDD))

	payload := fmt.Sprintf(`{ "jql":"%v", "startAt":0, "fields":["key"] }`, jql)

	var jsonStr = []byte(payload)
	r, _ := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonStr))
	r.SetBasicAuth(c.Username, c.Password)
	r.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return end, 0, err
	}
	defer resp.Body.Close()

	var jsr JiraSearchResponse
	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &jsr)
	if err != nil {
		return end, 0, err
	}
	return end, jsr.Total, nil
}
