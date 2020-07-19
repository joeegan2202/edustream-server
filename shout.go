package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var messagePoller *sql.Stmt
var messageClearer *sql.Stmt
var messagePoster *sql.Stmt

func pollShout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var lastID int

	if query["session"] == nil || query["lastID"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	lastID, err := strconv.Atoi(query["lastID"][0])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect format for last message id!"}`))
		return
	}

	if _, err := checkSession(sid, session); err != nil {
		logger.Printf("Error in pollShout trying to check session! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	_, err = messageClearer.Exec()

	if err != nil {
		logger.Printf("Error in pollShout trying to clear old messages! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Could not clear old messages!"}`)))
		return
	}

	for {
		wait := time.After(10 * time.Second)

		rows, err := messagePoller.Query(sid, lastID, session, sid)

		if err != nil {
			logger.Printf("Error trying to poll for messages! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error trying to poll for messages!")
			return
		}

		jsonAccumulator := "["

		send := false

		for rows.Next() {
			send = true
			var (
				id   uint64
				text string
			)

			err := rows.Scan(&id, &text)
			if err != nil {
				logger.Printf("Error trying to scan rows for messages! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error trying to scan rows for messages!")
				rows.Close()
				return
			}

			if jsonAccumulator == "[" {
				jsonAccumulator += fmt.Sprintf(`{"id": %d, "body": %s}`, id, text)
			} else {
				jsonAccumulator += fmt.Sprintf(`,{"id": %d, "body": %s}`, id, text)
			}
		}

		jsonAccumulator += "]"

		if send {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "shouts": %s}`, jsonAccumulator)))
			rows.Close()
			return
		}

		rows.Close()

		<-wait
	}
}

func postShout(w http.ResponseWriter, r *http.Request) {
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

	var session string
	var sid string

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]

	if _, err := checkSession(sid, session); err != nil {
		logger.Printf("Error in postShout trying to check session! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	text, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logger.Printf("Error trying to read message body! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error trying to read message body!")
		return
	}

	_, err = messagePoster.Exec(sid, text, session, sid)

	if err != nil {
		logger.Printf("Error trying to post message! %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error trying to post message!")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": ""}`)))
	return
}
