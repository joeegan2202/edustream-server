package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func teacherRoster() {
	var filename string
	fmt.Print("Enter the filename: ")
	fmt.Scanln(&filename)
	var hostname string
	fmt.Print("Enter the hostname: ")
	fmt.Scanln(&hostname)
	var username string
	fmt.Print("Enter the username: ")
	fmt.Scanln(&username)
	var password string
	fmt.Print("Enter the password: ")
	fmt.Scanln(&password)
	var port string
	fmt.Print("Enter the port: ")
	fmt.Scanln(&port)
	var sid string
	fmt.Print("Enter the school id: ")
	fmt.Scanln(&sid)

	file, err := os.Open(filename)

	if err != nil {
		log.Fatalf("Error opening file for reading! %s\n", err.Error())
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/edustream", username, password, hostname, port))

	if err != nil {
		log.Fatalf("Error trying to open database! %s\n", err.Error())
	}

	dataSheet := csv.NewReader(file)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 2)
	values, err := dataSheet.Read()
	if err != nil {
		log.Fatalf("Error trying to read first line of file! %s\n", err.Error())
	}

	for i, value := range values {
		switch value {
		case "id":
			indices[0] = i
		case "teachername":
			indices[1] = i
		}
	}

	tx, err := db.Begin()

	if err != nil {
		log.Fatalf("Error starting the transaction! %s\n", err.Error())
	}

	selectq, err := tx.Prepare("SELECT * FROM roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;")
	update, err := tx.Prepare("UPDATE roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid SET roster.cid=cnew.id WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;")
	insert, err := tx.Prepare("INSERT INTO roster VALUES ( ?, ?, ? );")
	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		names := strings.Split(record[indices[1]], ",")

		if len(names) != 2 {
			log.Println("Error trying to get names! Length not 2!")
			continue
		}

		var pid string

		err = db.QueryRow("SELECT id FROM people WHERE sid=? AND fname=? AND lname=?;", sid, strings.TrimSpace(names[1]), names[0]).Scan(&pid)

		if err != nil {
			log.Printf("Error trying to scan database for personid with name: %s. %s\n", record[indices[1]], err.Error())
		}

		rows, err := selectq.Query(sid, record[indices[0]], pid)

		if err != nil {
			log.Printf("Error while trying to query database to stop duplicates! pid: %s, cid: %s; %s\n", pid, record[indices[0]], err.Error())
		}

		defer rows.Close()

		if !rows.Next() {
			updated, err := update.Exec(sid, record[indices[0]], pid)

			if err != nil {
				log.Printf("Error while trying to update database for import! pid: %s, cid: %s; %s\n", pid, record[indices[0]], err.Error())
			}

			if num, _ := updated.RowsAffected(); num == 0 {
				_, err = insert.Exec(sid, pid, record[indices[0]])

				if err != nil {
					log.Printf("Error trying to insert rows while importing roster! pid: %s, cid: %s; %s\n", pid, record[indices[0]], err.Error())
				}
			}
		}

		rows.Close()
	}

	tx.Commit()
}
