package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func isError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func httpRead(res *http.Response) []byte {
	data, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	isError(err)
	return data
}

func httpGet(url string) *http.Response {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", url, nil)
	res, err := client.Do(req)
	isError(err)
	return res
}

func parseHTML(html string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	isError(err)
	return doc
}
