package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func addSession(uname string, sid string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%d", uname, time.Now().Unix())))
	id := fmt.Sprintf("%x", hash.Sum(nil))

	db.Exec("DELETE FROM sessions WHERE uname=? AND sid=?;", uname, sid)

	db.Exec("INSERT INTO sessions VALUES ( ?, ?, ?, ?);", sid, id, time.Now().Unix(), uname)

	return id
}

func passAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		sid   string
		uname string
	)

	if query["uname"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters!"}`))
		return
	}

	uname = query["uname"][0]
	sid = query["sid"][0]

	pword, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logger.Printf("Error trying to get password from request body! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to get password from request body!"}`)
		return
	}

	rows, err := db.Query("SELECT auth.password FROM auth INNER JOIN people ON people.id=auth.pid WHERE people.uname=? AND auth.sid=?;", uname, sid)

	if !rows.Next() {
		if err != nil {
			logger.Printf("Error trying to get password from database! %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to get password from database!"}`)
		return
	}

	var dbpass string

	err = rows.Scan(&dbpass)

	if err != nil {
		logger.Printf("Error trying to scan password from database! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to scan password from database!"}`)
		return
	}

	hash := sha256.New()
	hash.Write([]byte(pword))

	if dbpass != string(hash.Sum(nil)) {
		logger.Printf("Error trying to match given password to database password! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Wrong password!"}`)
		return
	}

	session := addSession(uname, sid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "session": "%s"}`, session)))
}
