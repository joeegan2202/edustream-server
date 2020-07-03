package main

import (
	"fmt"
	"log"
  "os"
	"net/http"
  "crypto/sha256"
	"github.com/gorilla/mux"
  "time"
)

var logger *log.Logger

func main() {
  f, err := os.OpenFile(fmt.Sprintf("logfile-%s.txt", time.Now().Format(time.RFC3339)), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    log.Fatal(err.Error())
  }
  defer f.Close()

  logger = log.New(f, "", log.Ldate | log.Ltime)

  hash := sha256.New()
  hash.Write([]byte("jeegan21Now"))
  sessionid := hash.Sum(nil)
  hash.Reset()
  hash.Write([]byte(fmt.Sprintf("jeegan21%d", time.Now().Unix())))
  uid := hash.Sum(nil)
  hash.Reset()
  hash.Write([]byte(fmt.Sprintf("admin%d", time.Now().Unix())))
  aid := hash.Sum(nil)
  hash.Reset()
  hash.Write([]byte(fmt.Sprintf("Spanish III X2016")))
  classid := hash.Sum(nil)
  hash.Reset()
  hash.Write([]byte(fmt.Sprintf("Cathedral High School")))
  sid := hash.Sum(nil)
  fmt.Printf("%x\n", sessionid)

  db = loadDatabase()

  createTables(db)

  db.Exec("DELETE FROM schools;")
  db.Exec("DELETE FROM sessions;") // Reset sessions and insert some dummy sessions for testing
  db.Exec("DELETE FROM roster;")
  db.Exec("DELETE FROM people;")
  db.Exec("DELETE FROM classes;")
  db.Exec("DELETE FROM periods;")
  db.Exec("INSERT INTO schools VALUES ( ?, 'http://home.eganshub.net', 'Cathedral High School', 'Indianapolis, IN', 'Insert public key here:' );", fmt.Sprintf("%x", sid))
  db.Exec("INSERT INTO people VALUES ( ?, ?, 'jeegan21', 'Joseph', 'Egan', 'S');", fmt.Sprintf("%x", sid), fmt.Sprintf("%x", uid))
  db.Exec("INSERT INTO people VALUES ( ?, ?, 'admin', 'Admin', 'Admin', 'A');", fmt.Sprintf("%x", sid), fmt.Sprintf("%x", aid))
  db.Exec("INSERT INTO classes VALUES ( ?, ?, 'Spanish III X', '2016', 'A');", fmt.Sprintf("%x", sid), fmt.Sprintf("%x", classid))
  db.Exec("INSERT INTO periods VALUES ( ?, 'A', ?, ?);", fmt.Sprintf("%x", sid), time.Date(2020, time.June, 17, 17, 0, 0, 0, time.Local).Unix(), time.Date(2020, time.July, 29, 22, 0, 0, 0, time.Local).Unix())
  db.Exec("INSERT INTO roster VALUES ( ?, ?, ?);", fmt.Sprintf("%x", sid), fmt.Sprintf("%x", uid), fmt.Sprintf("%x", classid))
  db.Exec("INSERT INTO sessions VALUES ( ?, ?, ?, 'jeegan21');", fmt.Sprintf("%x", sid), fmt.Sprintf("%x", sessionid), time.Now().Unix())
  db.Exec("INSERT INTO sessions VALUES ( ?, ?, ?, 'admin');", fmt.Sprintf("%x", sid), "91c39dbc8b36cfaeba98ca25ef56de400d1401f0d4dd6b4e0a081d4ed12e2af2", time.Now().Unix())

  r := mux.NewRouter()
  r.HandleFunc("/admin/start/camera/", adminStartCamera) // Admins can start and stop cameras
  r.HandleFunc("/admin/stop/camera/", adminStopCamera)
  r.HandleFunc("/admin/create/camera/", adminCreateCamera) // CRUD operations for camera management
  r.HandleFunc("/admin/read/camera/", adminReadCameras)
  r.HandleFunc("/admin/update/camera/", adminUpdateCamera)
  r.HandleFunc("/admin/delete/camera/", adminDeleteCamera)
  r.HandleFunc("/auth/", tempAuthorize)
  r.HandleFunc("/status/", receiveStatus)
  r.HandleFunc("/request/", requestStream) // For admins/teachers/students who are requesting a video stream
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer))) // The actual file server for streams
  r.PathPrefix("/ingest/").Handler(http.StripPrefix("/ingest/", new(IngestServer))) // The actual file server for streams
  logger.Fatal(http.ListenAndServe(":8080", r))
}

func getSchools(w http.ResponseWriter, r *http.Request) {
  query := r.URL.Query()

  if query["term"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing query parameters!"}`))
    return
  }


}
