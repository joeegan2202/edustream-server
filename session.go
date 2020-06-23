package main

import (
  "fmt"
  "net/http"
  "crypto/sha256"
  "time"
)

func addSession(uname string, sid string) string {
  hash := sha256.New()
  hash.Write([]byte(fmt.Sprintf("%s%d", uname, time.Now().Unix())))
  id := string(hash.Sum(nil))

  db.Exec("DELETE FROM sessions WHERE uname=? AND sid=?;", uname, sid)

  db.Exec("INSERT INTO sessions VALUES ( ?, ?, ?, ?);", sid, id, time.Now().Unix(), uname)

  return id
}

func tempAuthorize(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Content-Type", "application/json")

  query := r.URL.Query()

  var (
    uname string
    sid string
  )

  if query["uname"] == nil || query["sid"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters!"}`))
    return
  }

  uname = query["uname"][0]
  sid = query["sid"][0]

  session := addSession(uname, sid)

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "session": "%s"}`, session)))
}
