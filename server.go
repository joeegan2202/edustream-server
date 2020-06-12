package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (c *Camera) test(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  c.initiate("rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa", "testStream")
  w.Write([]byte(`{"message": "Stream started"}`))
}

func (c *Camera) resolveClosed(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  for true {
    select {
    case msg := <-c.stream.done:
      w.Write([]byte(fmt.Sprintf(`{"message": "%v"}`, msg)))
      break
    case msg := <-c.record.done:
      w.Write([]byte(fmt.Sprintf(`{"message": "%v"}`, msg)))
      break
    }
  }
}

type StreamServer struct {}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  http.FileServer(http.Dir("./testStream/")).ServeHTTP(w, r)
}

func main() {
  db := loadDatabase()

  rows, err := db.Query("SELECT * FROM roster;")
  
  if err != nil {
    log.Fatal(err.Error())
  }
  defer rows.Close()

  for rows.Next() {
    var pid int
    var cid int

    if err := rows.Scan(&pid, &cid); err != nil {
      log.Fatal(fmt.Sprintf("Error: %s", err))
    }
    fmt.Printf("PID: %d, CID: %d\n", pid, cid)
  }

  testCam := new(Camera)
  testCam.stream = new(Stream)
  testCam.record = new(Record)

  r := mux.NewRouter()
  r.HandleFunc("/", testCam.test)
  r.HandleFunc("/closed", testCam.resolveClosed)
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer)))
  log.Fatal(http.ListenAndServe(":8080", r))
}

