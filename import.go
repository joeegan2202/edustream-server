package main

import (
  "net/http"
  "encoding/csv"
  "fmt"
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
    sid string
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
    sid string
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
    sid string
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

  // Write rest of data to db
  for {
    record, err := dataSheet.Read()
    if err != nil {
      break
    }

    rows, err := db.Query("SELECT * FROM roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;", sid, record[indices[1]], record[indices[0]])

    if err != nil {
      logger.Printf("Error while trying to query database for import! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
      return
    }

    if !rows.Next() {
      updated, err := db.Exec("UPDATE roster INNER JOIN classes AS cold ON roster.cid=cold.id INNER JOIN classes AS cnew ON roster.sid=cnew.sid SET roster.cid=cnew.id WHERE roster.sid=? AND cnew.id=? AND roster.pid=? AND cold.period=cnew.period;", sid, record[indices[1]], record[indices[0]])

      if err != nil {
        logger.Printf("Error while trying to update database for import! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"status": false, "err": "Error trying to update database with records!"}`))
        return
      }

      if num, _ := updated.RowsAffected(); num == 0 {
        _, err = db.Exec("INSERT INTO roster VALUES ( ?, ?, ? );", sid, record[indices[0]], record[indices[1]])

        if err != nil {
          logger.Printf("Error trying to insert rows while importing roster! pid: %s, cid: %s; %s\n", record[indices[0]], record[indices[1]], err.Error())
          w.WriteHeader(http.StatusInternalServerError)
          w.Write([]byte(`{"status": false, "err": "Error trying to import roster!"}`))
          return
        }
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
    sid string
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

  _, err = db.Exec("DELETE FROM periods WHERE sid=?", sid)

  if err != nil {
    logger.Printf("Error trying to clear schedule! %s\n", err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Error trying to clear schedule!"}`))
    return
  }

  // Write rest of data to db
  for {
    record, err := dataSheet.Read()
    if err != nil {
      break
    }

    _, err = db.Exec("INSERT INTO periods VALUES ( ?, ?, ?, ? );", sid, record[indices[0]], record[indices[1]], record[indices[2]])

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
			id string
			uname string
      fname string
      lname string
      role string
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
			id string
			name string
      room string
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

  rows, err := db.Query("SELECT code, stime, etime FROM periods WHERE sid=?;", sid)
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
			code string
			stime uint64
      etime uint64
		)

		if err := rows.Scan(&code, &stime, &etime); err != nil {
      logger.Printf("Error in adminReadPeriods trying to scan row for period values! Error: %s\n", err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for period values"}`)))
      return
		}

    if jsonAccumulator != "[" {
      jsonAccumulator += ","
    }

    jsonAccumulator += fmt.Sprintf(`{"code": "%s", "stime": %d, "etime": %d}`, code, stime, etime)
	}

  jsonAccumulator += "]"

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "periods": %s }`, jsonAccumulator)))
  return
}

