package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	r := ranking()
	output := strings.Join(r, "\n")
	ioutil.WriteFile("output.csv", []byte(output), 0644)
}

func ranking() []string {
	rand.Seed(time.Now().UnixNano())
	header := "name, bio, country, Kaggle, Twitter, LinkedIn, Github, Blog"
	results := []string{header}

	for i := 0; i < 25; i++ {
		ranking, err := goquery.NewDocument("http://localhost:8050/render.html?url=https%3A%2F%2Fwww.kaggle.com%2Frankings.json%3Fgroup%3Dcompetitions%26page%3D" + fmt.Sprint(i+1) + "%26pageSize%3D20&timeout=10&wait=5")
		if err != nil {
			fmt.Println(err)
		}
		r := ranking.Text()
		var decodeData interface{}
		_ = json.Unmarshal([]byte(r), &decodeData)
		d := decodeData.(map[string]interface{})
		if d["list"] == nil {
			fmt.Println("Skip because of d[list] is nil.")
			continue
		}
		list := d["list"].([]interface{})
		for _, rank := range list {
			userURL := rank.(map[string]interface{})["userUrl"]
			result := user(fmt.Sprintf("https://www.kaggle.com%s", userURL))
			if result != "" {
				results = append(results, result)
			}
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		}
	}
	fmt.Printf("%v\n", results)
	return results
}

func user(url string) string {
	user, err := goquery.NewDocument(fmt.Sprintf("http://localhost:8050/render.html?url=%s&timeout=10&wait=5", url))
	if err != nil {
		fmt.Printf("NewDocument error. %v\n", err)
		return ""
	}
	s := user.Find("body > main > div > div.site-layout__main-content > script").Text()
	if len(strings.Split(s, "Kaggle.State.push(")) < 2 {
		fmt.Println("Skip because of insufficient data.")
		return ""
	}
	s = strings.Split(strings.Split(s, "Kaggle.State.push(")[1], ");")[0]
	var decodeData interface{}
	_ = json.Unmarshal([]byte(s), &decodeData)
	d := decodeData.(map[string]interface{})
	country := fmt.Sprintf("%s", d["country"])
	if country != "Japan" && country != "JP" {
		fmt.Printf("Skip because of not a Japanese. %s\n", country)
		return ""
	}
	result := fmt.Sprintf("%v, \"%v\", %v, %v, %v, %v, %v, %v",
		d["displayName"],
		strings.Replace(strings.Replace(fmt.Sprint(d["bio"]), "\n", " ", -1), ",", ";", -1), d["country"], url,
		fmt.Sprintf("https://twitter.com/%s", d["twitterUserName"]),
		d["linkedInUrl"],
		fmt.Sprintf("https://github.com/%s", d["gitHubUserName"]),
		d["websiteUrl"])

	fmt.Printf("%s\n", d["displayName"])
	fmt.Printf("%s\n", d["country"])
	fmt.Printf("%s\n", d["linkedInUrl"])
	fmt.Printf("%s\n", d["gitHubUserName"])
	fmt.Printf("%s\n", d["twitterUserName"])
	fmt.Printf("%s\n", d["websiteUrl"])
	fmt.Printf("%s\n", d["organization"])
	fmt.Printf("%s\n", d["bio"])
	fmt.Printf("%s\n", d["userAvatarUrl"])
	fmt.Printf("%s\n", d["email"])
	return result
}
