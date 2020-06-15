package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var (
  db *sql.DB
)

func checkSession(id string) (string, error) {
  rows, err := db.Query("SELECT time, uname FROM sessions WHERE id=?;", id)

  if err != nil {
    fmt.Println("Error getting first query")
    return "", err
  }

  fmt.Printf("Time: %d\n", time.Now().Unix())

  rows.Next()

  var sessionTime int64
  var uname string

  if err := rows.Scan(&sessionTime, &uname); err != nil {
    fmt.Println("Error doing the first scan")
    return "", err
  }

  fmt.Printf("Session time: %d\n", sessionTime)
  now := time.Now().Unix()
  if sessionTime < now - (30*60) {
    db.Exec("DELETE FROM sessions WHERE id=?;", id)
    return "", fmt.Errorf("Session: %s, too old! %d seconds too old.", id, now - sessionTime - (30*60))
  }

  rows.Close()

  row := db.QueryRow("SELECT role FROM people WHERE uname=?;", uname)

  var role string

  err = row.Scan(&role)

  if err != nil {
    fmt.Println("Error getting uname")
    return "", err
  }

  return role, nil
}

func loadDatabase() *sql.DB {
  if godotenv.Load("credentials.env") != nil {
    log.Fatal("Failed to get credentials while loading database")
  }

  uname := os.Getenv("DB_USER")
  pword := os.Getenv("DB_PASS")

  fmt.Printf("Username: %s, Password: %s\n", uname, pword)

  db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/classroom", uname, pword))

  if err != nil {
    log.Fatal(err.Error())
  }

  return db
}

func createTables(db *sql.DB) {
  db.Exec("CREATE TABLE cameras ( id CHAR(64) NOT NULL, address VARCHAR(255) NOT NULL, room VARCHAR(20) NOT NULL, framerate INT NOT NULL, bitrate VARCHAR(10) NOT NULL, hlsTime INT NOT NULL, hlsWrap INT NOT NULL, codec VARCHAR(20) NOT NULL, PRIMARY KEY (id) );")
  db.Exec("CREATE TABLE people ( id CHAR(64) NOT NULL, uname VARCHAR(20) NOT NULL, fname VARCHAR(20) NOT NULL, lname VARCHAR(20) NOT NULL, role CHAR NOT NULL, PRIMARY KEY(id) );")
  db.Exec("CREATE TABLE classes ( id CHAR(64) NOT NULL, name VARCHAR(40) NOT NULL, room VARCHAR(20) NOT NULL, period VARCHAR(10) NOT NULL, PRIMARY KEY(id) );")
  db.Exec("CREATE TABLE roster ( pid CHAR(64) NOT NULL, cid CHAR(64) NOT NULL, FOREIGN KEY (pid) REFERENCES people(id), FOREIGN KEY (cid) REFERENCES classes(id) );")
  db.Exec("CREATE TABLE periods ( code VARCHAR(10) NOT NULL, stime INT NOT NULL, etime INT NOT NULL, date DATE NOT NULL );")
  db.Exec("CREATE TABLE sessions ( id CHAR(64) NOT NULL, time INT NOT NULL, uname VARCHAR(20) NOT NULL, PRIMARY KEY (id) );")
}

func populateSomeData(db *sql.DB) {
  db.Exec("INSERT INTO cameras VALUES ( '84257', 'rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa', '4103', 30, '16M', 3, 10, 'copy' );")
  db.Exec("INSERT INTO people VALUES ( '18427', 'jeegan21', 'Joseph', 'Egan', 'S' );")
  db.Exec("INSERT INTO people VALUES ( '659244', 'mtegan22', 'Max', 'Egan', 'S' );")
  db.Exec("INSERT INTO people VALUES ( '472662', 'regan', 'Rose', 'Egan', 'T' );")
  db.Exec("INSERT INTO classes VALUES ( '88231', 'Spanish III X', '3301', 'A' );")
  db.Exec("INSERT INTO roster VALUES ( '18427', '88231' );")
  db.Exec("INSERT INTO roster VALUES ( '659244', '88231' );")
  db.Exec("INSERT INTO roster VALUES ( '472662', '88231' );")
  db.Exec("INSERT INTO periods VALUES ( 'A', 1300, 1400, 2020-06-15 );")
}
