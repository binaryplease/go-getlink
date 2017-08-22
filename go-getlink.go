package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"strings"
)

type Url struct {
	url string
	id  int
}

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

func save_links(urls []string, rootId int) {

}

func main() {

	reset_all_urls()

	batchsize := 5

	for {
		foundUrls := make(map[string]bool)
		seedUrls := get_urls(batchsize)

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

		// We're done! Print the results...

		fmt.Println("\nFound", len(foundUrls), "unique urls:")

		// Display results
		//for url, _ := range foundUrls {
		//fmt.Println(" - " + url)
		//}

		close(chUrls)

		save_urls(foundUrls)
		mark_urls_done(seedUrls)
	}
}

func save_urls(urls map[string]bool) {

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)
	stmt, err := db.Prepare("insert into urlresults (id,url) values (null, ? )")
	checkErr(err)

	for url, _ := range urls {
		_, err := stmt.Exec(url)
		checkErr(err)
	}
}

func set_url_status(urls []Url, status int) {

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)
	stmt, err := db.Prepare("update domains set status=? where id=?")
	checkErr(err)

	for _, url := range urls {
		_, err := stmt.Exec(status, url.id)
		checkErr(err)

	}
}

func reset_all_urls() {

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)
	stmt, err := db.Prepare("update domains set status=0")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)

}

func mark_urls_done(urls []Url) {
	set_url_status(urls, 1)
}

func reset_urls(urls []Url) {
	set_url_status(urls, 0)
}

func get_urls(num int) []Url {
	var urls []Url

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)
	rows, err := db.Query("SELECT * FROM domains where status = 0 limit " + strconv.Itoa(num))
	checkErr(err)

	var id int
	var url string
	var status int

	for rows.Next() {
		err = rows.Scan(&id, &url, &status)
		checkErr(err)
		urls = append(urls, Url{url, id})
	}

	db.Close()
	return urls
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
