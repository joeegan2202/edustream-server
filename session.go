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
		uname string
	)

	if query["uname"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters!"}`))
		return
	}

	uname = query["uname"][0]

	pword, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logger.Printf("Error trying to get password from request body! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to get password from request body!"}`)
		return
	}

	rows, err := db.Query("SELECT auth.password, people.role, schools.id, schools.banner FROM auth INNER JOIN people ON people.id=auth.pid INNER JOIN schools ON schools.id=auth.sid WHERE people.uname=?;", uname)

	if err != nil {
		logger.Printf("Error trying to get password from database! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to get password from database!"}`)
		return
	}
	if !rows.Next() {
		logger.Printf("Error trying to get password from database! No rows!\n")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to get password from database!"}`)
		return
	}

	defer rows.Close()

	var dbpass string
	var role string
	var sid string
	var bannerURL string

	err = rows.Scan(&dbpass, &role, &sid, &bannerURL)

	if err != nil {
		logger.Printf("Error trying to scan password from database! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Error trying to scan password from database!"}`)
		return
	}

	hash := sha256.New()
	hash.Write([]byte(pword))
	hashed := fmt.Sprintf("%x", hash.Sum(nil))

	logger.Printf("Authenticating %s with hashed pass: %s, and dbpass: %s\n", uname, hashed, dbpass)

	if dbpass != hashed {
		logger.Println("Error trying to match given password to database password!")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status": false, "err": "Wrong password!"}`)
		return
	}

	session := addSession(uname, sid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "session": "%s", "role": "%s", "sid": "%s", "bannerURL": "%s"}`, session, role, sid, bannerURL)))
}
