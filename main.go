package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := "https://google.co.jp"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(byteArray))

	doc, _ := goquery.NewDocument(url)
	fmt.Println(doc.Find("title").Text())
	user, _ := goquery.NewDocument("http://localhost:8050/render.html?url=https://www.kaggle.com/titericz&timeout=10&wait=5")
	fmt.Println(user.Find(".site-layout__main-content").Text())
}
