package main

import (
  "net/http"
  "encoding/csv"
  "fmt"
)

func importPeople(w http.ResponseWriter, r *http.Request) {
  if r.Method == "OPTIONS" {
    w.Header().Set("Access-Control-Allow-Origin", "*")
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
    case "uname":
      indices[0] = i
    case "fname":
      indices[1] = i
    case "lname":
      indices[2] = i
    case "role":
      indices[3] = i
    }
  }

  // Write rest of data to db
  for record, err := dataSheet.Read(); err == nil; {
    fmt.Println(record[indices[0]], record[indices[1]], record[indices[2]], record[indices[3]])
  }
}
