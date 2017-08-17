package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Domain struct {
	root     string
	subpages []string
}

func main() {
	domain := get_domain()
	get_links(&domain)
	show(domain)
}

func show(domain Domain) {
	fmt.Print("Domain root: ")
	fmt.Println(domain.root)
	fmt.Print("Domain's subpages: ")
	fmt.Println(domain.subpages)
}

func get_links(domain *Domain) {

}
func get_domain() Domain {

	db, err := sql.Open("mysql", "root:hallo@/gogetlink?charset=utf8")
	checkErr(err)

	rows, err := db.Query("SELECT * FROM domains where status = 0 limit 1")
	checkErr(err)

	var id int
	var domain string
	var status int
	var sub []string

	for rows.Next() {

		err = rows.Scan(&id, &domain, &status)
		checkErr(err)
		//fmt.Println(id)
		//fmt.Println(domain)
		//fmt.Println(status)
	}

	db.Close()

	return Domain{domain, sub}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
