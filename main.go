package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	user, _ := goquery.NewDocument("http://localhost:8050/render.html?url=https://www.kaggle.com/confirm&timeout=10&wait=5")
	s := user.Find("body > main > div > div.site-layout__main-content > script").Text()
	s = strings.Split(strings.Split(s, "Kaggle.State.push(")[1], ");")[0]
	fmt.Println(s[len(s)-30 : len(s)])
	var decodeData interface{}
	_ = json.Unmarshal([]byte(s), &decodeData)
	d := decodeData.(map[string]interface{})
	fmt.Printf("%s\n", d["country"])
	fmt.Printf("%s\n", d["linkedInUrl"])
	fmt.Printf("%s\n", d["gitHubUserName"])
	fmt.Printf("%s\n", d["twitterUserName"])

}
