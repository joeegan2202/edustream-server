package main

import (
	"fmt"
	"log"
	"net/http"
  "crypto/sha256"
	"github.com/gorilla/mux"
  "time"
)

func main() {
  hash := sha256.New()
  hash.Write([]byte("jeegan21Now"))
  jeegan21 := hash.Sum(nil)
  fmt.Printf("%x\n", jeegan21)

  db = loadDatabase()

  //createTables(db)
  //populateSomeData(db)

  db.Exec("DELETE FROM sessions;")
  db.Exec("INSERT INTO sessions VALUES ( ?, ?, 'jeegan21');", fmt.Sprintf("%x", jeegan21), time.Now().Unix())
  db.Exec("INSERT INTO sessions VALUES ( ?, ?, 'admin');", "91c39dbc8b36cfaeba98ca25ef56de400d1401f0d4dd6b4e0a081d4ed12e2af2", time.Now().Unix())

  role, err := checkSession("91c39dbc8b36cfaeba98ca25ef56de400d1401f0d4dd6b4e0a081d4ed12e2af2")

  if err != nil {
    log.Println(err.Error())
  }

  fmt.Printf("Role found: %v\n", role)

  r := mux.NewRouter()
  r.HandleFunc("/admin/start/camera/", adminStartCamera)
  r.HandleFunc("/admin/stop/camera/", adminStopCamera)
  r.HandleFunc("/admin/create/camera/", adminCreateCamera)
  r.HandleFunc("/admin/read/camera/", adminReadCameras)
  r.HandleFunc("/admin/update/camera/", adminUpdateCamera)
  r.HandleFunc("/admin/delete/camera/", adminDeleteCamera)
  r.HandleFunc("/request/", requestStream)
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer)))
  log.Fatal(http.ListenAndServe(":8080", r))
}

