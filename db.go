package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
	db *sql.DB
)

func checkSession(sid, session string) (string, error) {
	rows, err := db.Query("SELECT time, uname FROM sessions WHERE sid=? AND id=?;", sid, session)

	if err != nil {
		return "", err
	}

	rows.Next()

	var sessionTime int64
	var uname string

	if err := rows.Scan(&sessionTime, &uname); err != nil {
		return "", err
	}

	defer rows.Close()

	now := time.Now().Unix()
	if sessionTime < now-(30*60) {
		db.Exec("DELETE FROM sessions WHERE sid=? AND id=?;", sid, session)
		return "", fmt.Errorf("session: %s, too old! %d seconds too old", session, now-sessionTime-(30*60))
	}

	row := db.QueryRow("SELECT role FROM people WHERE sid=? AND uname=?;", sid, uname)

	var role string

	err = row.Scan(&role)

	if err != nil {
		return "", err
	}

	return role, nil
}

func loadDatabase() *sql.DB {
	if godotenv.Load("credentials.env") != nil {
		logger.Fatal("Failed to get credentials while loading database")
	}

	uname := os.Getenv("DB_USER")
	pword := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/edustream", uname, pword, host, port))

	if err != nil {
		logger.Fatal(err.Error())
	}

	return db
}

func createTables(db *sql.DB) {
	db.Exec("CREATE TABLE schools ( id CHAR(64) NOT NULL, address VARCHAR(255) NOT NULL, name VARCHAR(60) NOT NULL, city VARCHAR(60) NOT NULL, banner VARCHAR(255) NOT NULL, publicKey TEXT(540) NOT NULL, PRIMARY KEY (id) );")
	db.Exec("CREATE TABLE cameras ( sid CHAR(64) NOT NULL, id CHAR(64) NOT NULL, address VARCHAR(255) NOT NULL, room VARCHAR(20) NOT NULL, lastStreamed INT NOT NULL, locked INT NOT NULL, PRIMARY KEY (id), FOREIGN KEY(sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE people ( sid CHAR(64) NOT NULL, id CHAR(64) NOT NULL, uname VARCHAR(20) NOT NULL, fname VARCHAR(20) NOT NULL, lname VARCHAR(20) NOT NULL, role CHAR NOT NULL, PRIMARY KEY(id), FOREIGN KEY(sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE classes ( sid CHAR(64) NOT NULL, id CHAR(64) NOT NULL, name VARCHAR(40) NOT NULL, room VARCHAR(20) NOT NULL, period VARCHAR(10) NOT NULL, PRIMARY KEY(id), FOREIGN KEY(sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE roster ( sid CHAR(64) NOT NULL, pid CHAR(64) NOT NULL, cid CHAR(64) NOT NULL, FOREIGN KEY (pid) REFERENCES people(id), FOREIGN KEY (cid) REFERENCES classes(id), FOREIGN KEY(sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE periods ( sid CHAR(64) NOT NULL, id INT NOT NULL AUTO_INCREMENT, code VARCHAR(10) NOT NULL, stime INT NOT NULL, etime INT NOT NULL, FOREIGN KEY(sid) REFERENCES schools(id), PRIMARY KEY(id) );")
	db.Exec("CREATE TABLE sessions ( sid CHAR(64) NOT NULL, id CHAR(64) NOT NULL, time INT NOT NULL, uname VARCHAR(20) NOT NULL, PRIMARY KEY (id), FOREIGN KEY(sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE usage ( sid CHAR(64) NOT NULL, bytes BIGINT NOT NULL, FOREIGN KEY (sid) REFERENCES schools(id) );")
	db.Exec("CREATE TABLE auth ( sid CHAR(64) NOT NULL, pid CHAR(64) NOT NULL, password CHAR(64) NOT NULL, FOREIGN KEY (sid) REFERENCES schools(id), FOREIGN KEY (pid) REFERENCES people(id) );")
	db.Exec("CREATE TABLE messages ( sid CHAR(64) NOT NULL, id INT NOT NULL AUTO_INCREMENT, room VARCHAR(20) NOT NULL, text TEXT NOT NULL, etime INT NOT NULL, FOREIGN KEY (sid) REFERENCES schools(id), PRIMARY KEY (id) );")
	db.Exec("CREATE TABLE recording ( sid CHAR(64) NOT NULL, id INT NOT NULL AUTO_INCREMENT, cid CHAR(64) NOT NULL, time INT NOT NULL, status TINYINT NOT NULL, FOREIGN KEY(sid) REFERENCES schools(id), FOREIGN KEY(cid) REFERENCES classes(id), PRIMARY KEY(id) );")
}
