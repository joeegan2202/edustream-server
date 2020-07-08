package main

import (
  "net/http"
  "fmt"
  "encoding/csv"
)

func importPeople(w http.ResponseWriter, r *http.Request) {
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


func importClasses(w http.ResponseWriter, r *http.Request) {
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

  fmt.Printf("%d, %d, %d, %d\n", indices[0], indices[1], indices[2], indices[3])

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

func importRoster(w http.ResponseWriter, r *http.Request) {
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

    rows, err := db.Query("SELECT * FROM roster WHERE sid=? AND id=?;", sid, record[indices[1]])

    if err != nil {
      logger.Printf("Error while trying to query database for import! %s\n", err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
      return
    }

    if rows.Next() {
      _, err := db.Exec("UPDATE classes SET name=?, room=?, period=? WHERE sid=? AND id=?;", record[indices[0]], record[indices[1]], record[indices[2]], sid, record[indices[4]])

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

func importPeriods(w http.ResponseWriter, r *http.Request) {
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

    rows, err := db.Query("SELECT * FROM classes WHERE sid=? AND id=?;", sid, record[indices[4]])

    if err != nil {
      logger.Printf("Error while trying to query database for import! %s\n", err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(`{"status": false, "err": "Error trying to query database with records!"}`))
      return
    }

    if rows.Next() {
      _, err := db.Exec("UPDATE classes SET name=?, room=?, period=? WHERE sid=? AND id=?;", record[indices[0]], record[indices[1]], record[indices[2]], sid, record[indices[4]])

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

