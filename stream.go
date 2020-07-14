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
