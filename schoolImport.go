package main

import (
  "net/http"
  "encoding/csv"
  "fmt"
)

func importPeople(w http.ResponseWriter, r *http.Request) {
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

  for record, err := dataSheet.Read(); err == nil; {
    fmt.Println(record[0])
  }
}
