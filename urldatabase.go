package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type Url struct {
	url string
	id  int
}

type UrlDatabase struct {
	db               *sql.DB
	sourceTable      string
	destinationTable string
}

func (udb *UrlDatabase) Close() {
	udb.db.Close()
}

func NewUrlDatabase() *UrlDatabase {

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)

	udb := UrlDatabase{db, "domains", "urlresults"}
	return &udb
}

func (udb *UrlDatabase) save_urls(urls map[string]bool) {

	stmt, err := udb.db.Prepare("insert into urlresults (id,url) values (null, ? )")
	checkErr(err)

	for url, _ := range urls {
		_, err := stmt.Exec(url)
		checkErr(err)
	}
}

func (udb *UrlDatabase) set_url_status(urls []Url, status int) {

	stmt, err := udb.db.Prepare("update domains set status=? where id=?")
	checkErr(err)

	for _, url := range urls {
		_, err := stmt.Exec(status, url.id)
		checkErr(err)

	}
}

func (udb *UrlDatabase) reset_all_urls() {

	stmt, err := udb.db.Prepare("update domains set status=0")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)

}

func (udb *UrlDatabase) mark_urls_done(urls []Url) {
	udb.set_url_status(urls, 1)
}

func (udb *UrlDatabase) reset_urls(urls []Url) {
	udb.set_url_status(urls, 0)
}

func (udb *UrlDatabase) get_urls(num int) []Url {
	var urls []Url

	rows, err := udb.db.Query("SELECT * FROM domains where status = 0 limit " + strconv.Itoa(num))
	checkErr(err)

	var id int
	var url string
	var status int

	for rows.Next() {
		err = rows.Scan(&id, &url, &status)
		checkErr(err)
		urls = append(urls, Url{url, id})
	}

	return urls
}
