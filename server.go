package main

import (
	"fmt"
	"log"
	"net/http"
  "crypto/sha256"
	"github.com/gorilla/mux"
)

type StreamServer struct {}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  http.FileServer(http.Dir("./streams/")).ServeHTTP(w, r)
}

func main() {
  hash := sha256.New()
  hash.Write([]byte("Test String\n"))
  fmt.Printf("%x\n", hash.Sum(nil))

  db = loadDatabase()

  createTables(db)
  populateSomeData(db)

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
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer)))
  log.Fatal(http.ListenAndServe(":8080", r))
}

