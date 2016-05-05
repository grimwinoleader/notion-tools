package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/notion-tools/utils"
)

type NotionConfig struct {
	Token string
	URL   string
}

type NotionResponse struct {
	Errors []string `json:"errors"`
	Status string   `json:"status"`
}

func (c *NotionConfig) Report(id string, date time.Time, val int) {

	payload := fmt.Sprintf(`{ "ingredient_id":"%v", "date":"%v", "value":%v }`,
		id, date.Format(utils.YYYYMMDD), val)
	fmt.Println(payload)

	var jsonStr = []byte(payload)
	r, _ := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonStr))
	r.Header.Set("Authorization", c.Token)
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
