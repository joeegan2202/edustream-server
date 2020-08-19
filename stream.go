package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// StreamServer Type for the server streaming files
type StreamServer struct{}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

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
			room string
		)

		err = rows.Scan(&sid, &room)

		if err != nil {
			logger.Printf("Error trying to scan database for stream server! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error trying to scan database for data!"))
			return
		}

		filename := strings.Split(r.URL.Path, "/")
		path := fmt.Sprintf("%s/%s/%s/%s", os.Getenv("FS_PATH"), sid, room, filename[len(filename)-1])

		//for _, cfile := range cache {
		//	if cfile.path == path {
		//		w.WriteHeader(http.StatusOK)
		//		w.Write(*cfile.data)
		//		return
		//	}
		//}

		for {
			file, err := os.Open(path)

			if err != nil {
				logger.Printf("Error opening file to serve stream! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Could not open file!"))
				return
			}

			//err = insertCache(path, io.TeeReader(file, w)) // Tries to read from file to both cache and http response. May have issues with latency while writing
			io.Copy(w, file)

			if err == nil {
				w.WriteHeader(http.StatusOK)
				file.Close()
				break
			}

			logger.Printf("Error trying to read file to insert into cache! %s\n", err.Error())
			file.Close()
		}
	case "A":
		filename := strings.Split(r.URL.Path, "/")
		path := fmt.Sprintf("%s/%s/%s/%s", os.Getenv("FS_PATH"), sid, filename[len(filename)-2], filename[len(filename)-1])

		//for _, cfile := range cache {
		//	if cfile.path == path {
		//		w.WriteHeader(http.StatusOK)
		//		w.Write(*cfile.data)
		//		return
		//	}
		//}

		for {
			file, err := os.Open(path)

			if err != nil {
				logger.Printf("Error opening file to serve stream! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Could not open file!"))
				return
			}

			//err = insertCache(path, io.TeeReader(file, w)) // Tries to read from file to both cache and http response. May have issues with latency while writing
			io.Copy(w, file)

			if err == nil {
				w.WriteHeader(http.StatusOK)
				file.Close()
				break
			}

			logger.Printf("Error trying to read file to insert into cache! %s\n", err.Error())
			file.Close()
		}
	}
}

// RecordServer Type for the server streaming the recorded files
type RecordServer struct{}

func (s *RecordServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	sid := strings.Split(r.URL.Path, "/")[0]
	session := strings.Split(r.URL.Path, "/")[1]
	rid := strings.Split(r.URL.Path, "/")[2]

	role, err := checkSession(sid, session)

	if err != nil {
		logger.Printf("Error checking sessions! %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error checking session!"))
		return
	}

	_, err = db.Exec("UPDATE sessions SET time=unix_timestamp() WHERE id=? AND sid=?;", session, sid)

	if err != nil {
		logger.Printf("Error in RecordServer trying to update session time! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error trying to update session time!"}`)))
		return
	}

	switch role {
	case "S", "T":
		rows, err := db.Query(`SELECT recording.cid, recording.time FROM recording
		INNER JOIN classes ON classes.id=recording.cid
		INNER JOIN roster ON roster.cid=classes.id
		INNER JOIN people ON roster.pid=people.id
		INNER JOIN sessions ON people.uname=sessions.uname
		WHERE recording.id=? AND sessions.sid=? AND sessions.id=?;`, rid, sid, session)

		if err != nil {
			logger.Printf("Error querying sessions to stream! %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error querying session to stream file!"))
			return
		}

		defer rows.Close()

		if !rows.Next() {
			logger.Printf("No recording found for id: %s\n", rid)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": false, "err": "No recording found!"}`))
			return
		}

		var (
			cid   string
			stime uint64
		)

		err = rows.Scan(&cid, &stime)

		if err != nil {
			logger.Printf("Error trying to scan database for record server! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error trying to scan database for data!"))
			return
		}

		recordTime := time.Unix(int64(stime), 0)

		filename := strings.Split(r.URL.Path, "/")
		path := fmt.Sprintf("%s/%s/%s-%d-%d/%s", sid, cid, recordTime.Month().String(), recordTime.Day(), recordTime.Year(), filename[len(filename)-1])

		object, err := minioClient.GetObject(context.Background(), "edustream-record", path, minio.GetObjectOptions{})

		if err != nil {
			logger.Printf("Error getting file from Spaces to serve stream! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not open file!"))
			return
		}

		_, err = io.Copy(w, object)

		if err == nil {
			w.WriteHeader(http.StatusOK)
			break
		}

		logger.Printf("Error trying to read file! %s\n", err.Error())
	case "A":
		rows, err := db.Query(`SELECT cid, time FROM recording WHERE id=? AND sid=?;`, rid, sid)

		if err != nil {
			logger.Printf("Error querying sessions to stream! %s\n", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error querying session to stream file!"))
			return
		}

		defer rows.Close()

		if !rows.Next() {
			logger.Printf("No recording found for id: %s\n", rid)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": false, "err": "No recording found!"}`))
			return
		}

		var (
			cid   string
			stime uint64
		)

		err = rows.Scan(&cid, &stime)

		if err != nil {
			logger.Printf("Error trying to scan database for record server! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error trying to scan database for data!"))
			return
		}

		recordTime := time.Unix(int64(stime), 0)

		filename := strings.Split(r.URL.Path, "/")
		path := fmt.Sprintf("%s/%s/%s-%d-%d/%s", sid, cid, recordTime.Month().String(), recordTime.Day(), recordTime.Year(), filename[len(filename)-1])

		object, err := minioClient.GetObject(context.Background(), "edustream-record", path, minio.GetObjectOptions{})

		if err != nil {
			logger.Printf("Error getting file from Spaces to serve stream! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not open file!"))
			return
		}

		_, err = io.Copy(w, object)

		if err == nil {
			w.WriteHeader(http.StatusOK)
			break
		}

		logger.Printf("Error trying to read file! %s\n", err.Error())
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
	case "A":
		if query["room"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
			return
		}

		room := query["room"][0]

		var temp string

		defer fmt.Printf("CID found: %s\n", temp)

		if db.QueryRow("SELECT classes.id FROM classes INNER JOIN periods ON classes.period=periods.code WHERE classes.room=? AND classes.sid=? AND periods.stime<unix_timestamp() AND periods.etime>unix_timestamp();", room, sid).Scan(&temp) != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "info": {"cname": "No class", "period": "No period", "attendance": []}}`)))
			return
		}

		row := db.QueryRow("SELECT classes.name, periods.code FROM classes INNER JOIN periods ON periods.code=classes.period WHERE periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() AND periods.sid=? AND classes.room=?;", sid, room)

		var (
			cname  string
			period string
		)

		err = row.Scan(&cname, &period)

		if err != nil {
			logger.Printf("Error in streamInfo trying to scan for information! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Error in streamInfo trying to scan for admin information!"}`)))
			return
		}

		rows, err := db.Query("SELECT people.fname, people.lname FROM people INNER JOIN sessions ON sessions.uname=people.uname INNER JOIN roster ON roster.pid=people.id INNER JOIN classes ON classes.id=roster.cid WHERE classes.id=( SELECT classes.id FROM classes INNER JOIN periods ON classes.period=periods.code WHERE classes.room=? AND classes.sid=? AND periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() ) AND sessions.time>unix_timestamp()-60;", room, sid)

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
		w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "info": {"cname": "%s", "period": "%s", "attendance": %s}}`, cname, period, jsonAccumulator)))
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
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "Error in streamInfo trying to scan for student information!", "info": {"fname": "No", "lname": "Name", "cname": "No Class", "period": "No Period"}}`)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "", "info": {"fname": "%s", "lname": "%s", "cname": "%s", "period": "%s"}}`, fname, lname, cname, period)))
	case "T":
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
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": true, "err": "Error in streamInfo trying to scan for teacher information!", "info": {"fname": "No", "lname": "Name", "cname": "No Class", "period": "No Period"}}`)))
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
