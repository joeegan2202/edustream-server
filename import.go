package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
)

func adminImportPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	dataSheet := csv.NewReader(r.Body)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 5)
	values, err := dataSheet.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not read header line from csv"}`))
		return
	}

	for i, value := range values {
		switch value {
		case "uname":
			indices[0] = i
		case "fname":
			indices[1] = i
		case "lname":
			indices[2] = i
		case "role":
			indices[3] = i
		case "id":
			indices[4] = i
		}
	}

	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		rows, err := db.Query("SELECT * FROM people WHERE sid=? AND id=?;", sid, record[indices[4]])

		if err != nil {
			logger.Printf("Error while trying to query database for import! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
			return
		}

		defer rows.Close()

		if rows.Next() {
			_, err := db.Exec("UPDATE people SET uname=?, fname=?, lname=?, role=? WHERE sid=? AND id=?;", record[indices[0]], record[indices[1]], record[indices[2]], record[indices[3]], sid, record[indices[4]])

			if err != nil {
				logger.Printf("Error while trying to update database for import! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to update database with records!"}`))
				return
			}
		} else {
			_, err = db.Exec("INSERT INTO people VALUES ( ?, ?, ?, ?, ?, ? );", sid, record[indices[4]], record[indices[0]], record[indices[1]], record[indices[2]], record[indices[3]])

			if err != nil {
				logger.Printf("Error trying to insert rows while importing people! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to import people!"}`))
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminImportClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	dataSheet := csv.NewReader(r.Body)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 4)
	values, err := dataSheet.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not read header line from csv"}`))
		return
	}

	for i, value := range values {
		switch value {
		case "name":
			indices[0] = i
		case "room":
			indices[1] = i
		case "period":
			indices[2] = i
		case "id":
			indices[3] = i
		}
	}

	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		rows, err := db.Query("SELECT * FROM classes WHERE sid=? AND id=?;", sid, record[indices[3]])

		if err != nil {
			logger.Printf("Error while trying to query database for import! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
			return
		}

		defer rows.Close()

		if rows.Next() {
			_, err := db.Exec("UPDATE classes SET name=?, room=?, period=? WHERE sid=? AND id=?;", record[indices[0]], record[indices[1]], record[indices[2]], sid, record[indices[3]])

			if err != nil {
				logger.Printf("Error while trying to update database for import! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to update database with records!"}`))
				return
			}
		} else {
			_, err = db.Exec("INSERT INTO classes VALUES ( ?, ?, ?, ?, ? );", sid, record[indices[3]], record[indices[0]], record[indices[1]], record[indices[2]])

			if err != nil {
				logger.Printf("Error trying to insert rows while importing classes! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to import classes!"}`))
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminImportRoster(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	dataSheet := csv.NewReader(r.Body)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 2)
	values, err := dataSheet.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not read header line from csv"}`))
		return
	}

	for i, value := range values {
		switch value {
		case "pid":
			indices[0] = i
		case "cid":
			indices[1] = i
		}
	}

	selectq, err := db.Prepare("SELECT * FROM roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;")
	update, err := db.Prepare("UPDATE roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid SET roster.cid=cnew.id WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;")
	insert, err := db.Prepare("INSERT INTO roster VALUES ( ?, ?, ? );")
	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		rows, err := selectq.Query(sid, record[indices[1]], record[indices[0]])

		if err != nil {
			logger.Printf("Error while trying to query database for import! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
			return
		}

		if !rows.Next() {
			updated, err := update.Exec(sid, record[indices[1]], record[indices[0]])

			if err != nil {
				logger.Printf("Error while trying to update database for import! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to update database with records!"}`))
				rows.Close()
				return
			}

			if num, _ := updated.RowsAffected(); num == 0 {
				_, err = insert.Exec(sid, record[indices[0]], record[indices[1]])

				if err != nil {
					logger.Printf("Error trying to insert rows while importing roster! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
				}
			}
		}

		rows.Close()
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminImportAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	dataSheet := csv.NewReader(r.Body)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 2)
	values, err := dataSheet.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not read header line from csv"}`))
		return
	}

	for i, value := range values {
		switch value {
		case "pid":
			indices[0] = i
		case "password":
			indices[1] = i
		}
	}

	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		rows, err := db.Query("SELECT * FROM auth WHERE sid=? AND pid=?;", sid, record[indices[0]])

		if err != nil {
			logger.Printf("Error while trying to query database for import! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
			return
		}

		defer rows.Close()

		password := sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(record[indices[1]])))))

		if rows.Next() {
			_, err := db.Exec("UPDATE auth SET password=? WHERE sid=? AND id=?;", password, sid, record[indices[3]])

			if err != nil {
				logger.Printf("Error while trying to update database for import! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to update database with records!"}`))
				return
			}
		} else {
			_, err = db.Exec("INSERT INTO auth VALUES ( ?, ?, ? );", sid, record[indices[0]], fmt.Sprintf("%x", password))

			if err != nil {
				logger.Printf("Error trying to insert rows while importing auth! %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": false, "err": "Error trying to import auth!"}`))
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminImportPeriods(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	dataSheet := csv.NewReader(r.Body)
	dataSheet.Comment = '#'

	// Get headers from csv
	indices := make([]int, 4)
	values, err := dataSheet.Read()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not read header line from csv"}`))
		return
	}

	for i, value := range values {
		switch value {
		case "code":
			indices[0] = i
		case "stime":
			indices[1] = i
		case "etime":
			indices[2] = i
		}
	}

	// Write rest of data to db
	for {
		record, err := dataSheet.Read()
		if err != nil {
			break
		}

		_, err = db.Exec("INSERT INTO periods (sid, code, stime, etime ) VALUES ( ?, ?, ?, ? );", sid, record[indices[0]], record[indices[1]], record[indices[2]])

		if err != nil {
			logger.Printf("Error trying to insert rows while importing periods! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to import periods!"}`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminReadPeople(w http.ResponseWriter, r *http.Request) {
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

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadPeople trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role for session"}`))
		return
	}

	rows, err := db.Query("SELECT id, uname, fname, lname, role FROM people WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadPeople querying database for people! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get people"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			id    string
			uname string
			fname string
			lname string
			role  string
		)

		if err := rows.Scan(&id, &uname, &fname, &lname, &role); err != nil {
			logger.Printf("Error in adminReadPeople trying to scan row for person values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for person values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"id": "%s", "uname": "%s", "fname": "%s", "lname": "%s", "role": "%s"}`, id, uname, fname, lname, role)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "people": %s }`, jsonAccumulator)))
	return
}

func adminReadClasses(w http.ResponseWriter, r *http.Request) {
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

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadClasses trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role for session"}`))
		return
	}

	rows, err := db.Query("SELECT id, name, room, period FROM classes WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadClasses querying database for classes! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get classes"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			id     string
			name   string
			room   string
			period string
		)

		if err := rows.Scan(&id, &name, &room, &period); err != nil {
			logger.Printf("Error in adminReadClasses trying to scan row for class values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for class values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"id": "%s", "name": "%s", "room": "%s", "period": "%s"}`, id, name, room, period)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "classes": %s }`, jsonAccumulator)))
	return
}

func adminReadRoster(w http.ResponseWriter, r *http.Request) {
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

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadRoster trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role for session"}`))
		return
	}

	rows, err := db.Query("SELECT pid, cid FROM roster WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadRoster querying database for roster! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get roster"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			pid string
			cid string
		)

		if err := rows.Scan(&pid, &cid); err != nil {
			logger.Printf("Error in adminReadRoster trying to scan row for roster values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for roster values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"pid": "%s", "cid": "%s"}`, pid, cid)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "roster": %s }`, jsonAccumulator)))
	return
}

func adminReadPeriods(w http.ResponseWriter, r *http.Request) {
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

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadPeriods trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role for session"}`))
		return
	}

	rows, err := db.Query("SELECT code, id, stime, etime FROM periods WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadPeriods querying database for periods! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get periods"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			code  string
			id    uint64
			stime uint64
			etime uint64
		)

		if err := rows.Scan(&code, &id, &stime, &etime); err != nil {
			logger.Printf("Error in adminReadPeriods trying to scan row for period values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for period values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"code": "%s", "id": "%d", "stime": %d, "etime": %d}`, code, id, stime, etime)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "periods": %s }`, jsonAccumulator)))
	return
}

func adminReadAuth(w http.ResponseWriter, r *http.Request) {
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

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadAuth trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role for session"}`))
		return
	}

	rows, err := db.Query("SELECT auth.pid, people.uname FROM auth INNER JOIN people ON people.id=auth.pid WHERE auth.sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadAuth querying database for authenticated people! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get authenticated people"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			pid   string
			uname string
		)

		if err := rows.Scan(&pid, &uname); err != nil {
			logger.Printf("Error in adminReadAuth trying to scan row for auth values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for auth values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"pid": "%s", "uname": "%s"}`, pid, uname)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "auth": %s }`, jsonAccumulator)))
	return
}

func adminUpdateAuth(w http.ResponseWriter, r *http.Request) {
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

	if query["session"] == nil || query["pid"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect parameters given!"}`))
		return
	}

	var (
		session string
		sid     string
		pid     string
	)

	session = query["session"][0]
	sid = query["sid"][0]
	pid = query["pid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error while checking session!"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Incorrect role"}`))
		return
	}

	pword, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error while reading from request body!"}`))
		return
	}

	hash := sha256.New()
	hash.Write([]byte(pword))

	rows, err := db.Query("SELECT * FROM auth WHERE sid=? AND pid=?;", sid, pid)

	if err != nil {
		logger.Printf("Error while scanning db for auth user!")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error while scanning db for auth users!"}`))
		return
	}

	defer rows.Close()

	if rows.Next() {
		_, err = db.Exec("UPDATE auth SET password=? WHERE sid=? AND pid=?;", fmt.Sprintf("%x", hash.Sum(nil)), sid, pid)

		if err != nil {
			logger.Printf("Error while trying to query database for import! %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
			return
		}
	} else {
		_, err = db.Exec("INSERT INTO auth VALUES ( ?, ?, ? );", sid, pid, fmt.Sprintf("%x", hash.Sum(nil)))

		if err != nil {
			logger.Printf("Error while trying to insert auth user!")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error while trying to insert auth user!"}`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminDeletePeople(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeletePeople checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM people WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeletePeople trying to query database for requested person! Error: %s\n", err.Error())
	}
	defer row.Close()
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Person does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM people WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeletePeople trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting person record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminDeleteClasses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeleteClasses checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM classes WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeleteClasses trying to query database for requested class! Error: %s\n", err.Error())
	}
	defer row.Close()
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Class does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM classes WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeleteClasses trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting class record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminDeleteRoster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeleteRoster checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM roster WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeleteRoster trying to query database for requested roster! Error: %s\n", err.Error())
	}
	defer row.Close()
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Roster does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM roster WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeleteRoster trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting roster record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminDeletePeriods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeletePeriods checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM periods WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeletePeriods trying to query database for requested period! Error: %s\n", err.Error())
	}
	defer row.Close()
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Period does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM periods WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeletePeriods trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting period record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminDeleteAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeleteAuth checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM auth WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeleteAuth trying to query database for requested auth! Error: %s\n", err.Error())
	}
	defer row.Close()
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Auth does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM auth WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeleteAuth trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting auth record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}
