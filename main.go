package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("Go MySQL Tutorial")

	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb
	db, err := sql.Open("mysql", "root:TrieIsHandsomeMan!@tcp(localhost:3306)/hand_cup")

	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT name FROM handcup_shop_info")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

}
