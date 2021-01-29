package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var outMutex *sync.Mutex
var errMutex *sync.Mutex

func printOut(s string, args ...interface{}) {
	outMutex.Lock()
	defer outMutex.Unlock()
	fmt.Fprintf(os.Stdout, s, args...)
}

func printErr(s string, args ...interface{}) {
	errMutex.Lock()
	defer errMutex.Unlock()
	fmt.Fprintf(os.Stderr, s, args...)
}

func getURLTitle(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		printErr("could not fetch title for url `%s`: %v\n", url, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		var bodyStr string
		if err != nil {
			bodyStr = fmt.Sprintf("could not parse body: %v", err)
		} else {
			bodyStr = string(body)
		}
		printErr("got a http response error for url `%s`: (%d %s) %s\n", url, resp.StatusCode, resp.Status, bodyStr)
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		printErr("could not start goquery for url `%s`: %v\n", url, err)
		return ""
	}

	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.Join(strings.Split(title, " "), " ")
	return title
}

func main() {
	separator := os.Getenv("OPTIMISTIC_TITLE_SEPARATOR")
	if separator == "" {
		separator = ","
	}

	errMutex = &sync.Mutex{}
	outMutex = &sync.Mutex{}

	s := bufio.NewScanner(os.Stdin)
	sem := make(chan struct{}, 500)
	wg := sync.WaitGroup{}
	for s.Scan() {
		sem <- struct{}{}
		wg.Add(1)
		go func(url string) {
			defer func() {
				<-sem
				wg.Done()
			}()
			printOut("%s%s%s\n", url, separator, getURLTitle(url))
		}(s.Text())
	}
	wg.Wait()
}
