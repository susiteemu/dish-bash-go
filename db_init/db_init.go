package dbinit

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

func Init() {

	os.Remove("./dishbashgo.db")

	db, err := sql.Open("sqlite3", "./dishbashgo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
	create table dish (id integer not null primary key, name text, url text, created date);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// generate some test data
	stmt, err := tx.Prepare("insert into dish(id, name, url, created) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("Ruoka %03d", i), fmt.Sprintf("https://ruoka%03d.fi", i), time.Now())
		if err != nil {
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

}
