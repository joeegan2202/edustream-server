package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type StreamServer struct{}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sid := strings.Split(r.URL.Path, "/")[0]
	session := strings.Split(r.URL.Path, "/")[1]

	now := time.Now().Unix()

	role, err := checkSession(sid, session)

	if err != nil {
		logger.Printf("Error checking sessions! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error checking session!"))
		return
	}

	_, err = db.Exec("UPDATE sessions SET time=unix_timestamp() WHERE id=? AND sid=?;", session, sid)

	if err != nil {
		logger.Printf("Error in streamInfo trying to update session time! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error trying to update session time!"}`)))
		return
	}

	switch role {
	case "S", "T":
		rows, err := db.Query(`SELECT sessions.sid, classes.room FROM sessions
    INNER JOIN people ON sessions.uname=people.uname
    INNER JOIN roster ON people.id=roster.pid
    INNER JOIN classes ON roster.cid=classes.id
    INNER JOIN periods ON classes.period=periods.code
    WHERE periods.stime<? AND periods.etime>? AND sessions.id=?;`, now, now, session)

		if err != nil {
			logger.Printf("Error querying sessions to stream! %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error querying session to stream file!"))
			return
		}

		defer rows.Close()

		if !rows.Next() {
			logger.Printf("No current period found for stream!")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": false, "err": "No session for stream found"}`))
			return
		}

		var (
			sid  string
			room uint64
		)

		err = rows.Scan(&sid, &room)

		fmt.Printf("About to serve %s/%s/%d\n", os.Getenv("FS_PATH"), sid, room)
		http.StripPrefix(fmt.Sprintf("%s/%s", sid, session), http.FileServer(http.Dir(fmt.Sprintf("%s/%s/%d", os.Getenv("FS_PATH"), sid, room)))).ServeHTTP(w, r)
	case "A":
		fmt.Printf("About to serve %s/%s\n", os.Getenv("FS_PATH"), sid)
		http.StripPrefix(fmt.Sprintf("%s/%s", sid, session), http.FileServer(http.Dir(fmt.Sprintf("%s/%s", os.Getenv("FS_PATH"), sid)))).ServeHTTP(w, r)
	}
}

func streamInfo(w http.ResponseWriter, r *http.Request) {
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

	role, err := checkSession(sid, session)

	if err != nil {
		logger.Printf("Error in streamInfo trying to check session! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	switch role {
	case "S":
		row := db.QueryRow("SELECT people.fname, people.lname, classes.name, periods.code FROM sessions INNER JOIN people ON sessions.uname=people.uname INNER JOIN roster ON people.id=roster.pid INNER JOIN classes ON roster.cid=classes.id INNER JOIN periods ON classes.period=periods.code WHERE periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() AND sessions.id=? AND sessions.sid=?;", session, sid)

		var (
			fname  string
			lname  string
			cname  string
			period string
		)

		err := row.Scan(&fname, &lname, &cname, &period)

		if err != nil {
			logger.Printf("Error in streamInfo trying to scan for information! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error in streamInfo trying to scan for information!"}`)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "info": {"fname": "%s", "lname": "%s", "cname": "%s", "period": "%s"}}`, fname, lname, cname, period)))
	case "T", "A":
		row := db.QueryRow("SELECT people.fname, people.lname, classes.name, periods.code FROM sessions INNER JOIN people ON sessions.uname=people.uname INNER JOIN roster ON people.id=roster.pid INNER JOIN classes ON roster.cid=classes.id INNER JOIN periods ON classes.period=periods.code WHERE periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() AND sessions.id=? AND sessions.sid=?;", session, sid)

		var (
			fname  string
			lname  string
			cname  string
			period string
		)

		err := row.Scan(&fname, &lname, &cname, &period)

		if err != nil {
			logger.Printf("Error in streamInfo trying to scan for information! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error in streamInfo trying to scan for information!"}`)))
			return
		}

		rows, err := db.Query("SELECT people.fname, people.lname FROM people INNER JOIN sessions ON sessions.uname=people.uname INNER JOIN roster ON roster.pid=people.id INNER JOIN classes ON classes.id=roster.cid WHERE classes.id=( SELECT classes.id FROM classes INNER JOIN roster on classes.id=roster.cid INNER JOIN people ON people.id=roster.pid INNER JOIN sessions ON sessions.uname=people.uname INNER JOIN periods ON classes.period=periods.code WHERE sessions.id=? AND people.sid=? AND periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() ) AND sessions.time>unix_timestamp()-60;", session, sid)

		if err != nil {
			logger.Printf("Error in streamInfo trying to scan for student attendance! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error in streamInfo trying to scan for student attendance!"}`)))
			return
		}

		defer rows.Close()

		jsonAccumulator := "["

		for rows.Next() {
			var (
				fname string
				lname string
			)

			err = rows.Scan(&fname, &lname)

			if err != nil {
				logger.Printf("Error in streamInfo trying to scan names for student attendance! Error: %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error in streamInfo trying to scan names for student attendance!"}`)))
				return
			}

			if jsonAccumulator == "[" {
				jsonAccumulator += fmt.Sprintf("\"%s %s\"", fname, lname)
			} else {
				jsonAccumulator += fmt.Sprintf(",\"%s %s\"", fname, lname)
			}
		}

		jsonAccumulator += "]"

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "info": {"fname": "%s", "lname": "%s", "cname": "%s", "period": "%s", "attendance": %s}}`, fname, lname, cname, period, jsonAccumulator)))
	}
}
