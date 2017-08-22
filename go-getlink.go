package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Extract all http** links from a given webpage
func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {

	udb := NewUrlDatabase()
	udb.reset_all_urls()

	batchsize := 5

	for {
		foundUrls := make(map[string]bool)
		seedUrls := udb.get_urls(batchsize)

		if len(seedUrls) == 0 {
			break
		}

		// Channels
		chUrls := make(chan string)
		chFinished := make(chan bool)

		// Kick off the crawl process (concurrently)
		for _, url := range seedUrls {
			go crawl(url.url, chUrls, chFinished)
		}

		// Subscribe to both channels
		for c := 0; c < len(seedUrls); {
			select {
			case url := <-chUrls:
				foundUrls[url] = true
			case <-chFinished:
				c++
			}
		}

		fmt.Println("\nFound", len(foundUrls), "unique urls")
		close(chUrls)

		udb.save_urls(foundUrls)
		udb.mark_urls_done(seedUrls)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
