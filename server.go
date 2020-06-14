package main

import (
	"fmt"
	"log"
	"net/http"
  "crypto/sha256"
	"github.com/gorilla/mux"
)

func (c *Camera) test(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  c.inputAddress = "rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa"
  c.outputFolder = "testStream"
  c.initiate()
  w.Write([]byte(`{"message": "Stream started"}`))
}

func (c *Camera) resolveClosed(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  for true {
    select {
    case msg := <-c.streamDone:
      w.Write([]byte(fmt.Sprintf(`{"message": "%v"}`, msg)))
      break
    case msg := <-c.recordDone:
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

func adminAddCamera(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Content-Type", "application/json")

  query := r.URL.Query()

  session := query["sessionid"]
  if session == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "No Session ID"}`))
    return
  } else if role, err := checkSession(session[0]); role != "A" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "%s"}`, err.Error())))
    return
  }
  address := query["address"]
  if address == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "No Address"}`))
    return
  }
  room := query["room"]
  if room == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "No Room Code"}`))
    return
  }
  framerate := query["framerate"]
  bitrate := query["bitrate"]
  hlsTime := query["hlsTime"]
  hlsWrap := query["hlsWrap"]
  codec := query["codec"]

  row := db.QueryRow("SELECT * FROM cameras WHERE address=? OR room=?;", address[0], room[0])
  if err := row.Scan(); err == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera with address or room code already created"}`)))
    return
  }

  _, err := db.Exec("INSERT INTO cameras VALUES (?, ?, ?, ?, ?, ?, ?, ?);", 928468721, address, room, framerate, bitrate, hlsTime, hlsWrap, codec)

  if err != nil {
    fmt.Printf("Error found: %s\n", err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Error creating camera record"}`))
    return
  }

  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(`{"status": true, "err": ""}`))

  return
}

func main() {
  hash := sha256.New()
  hash.Write([]byte("Test String\n"))
  fmt.Printf("%x\n", hash.Sum(nil))

  db = loadDatabase()

  //createTables(db)
  //populateSomeData(db)

  testCam := new(Camera)

  role, err := checkSession("91c39dbc8b36cfaeba98ca25ef56de400d1401f0d4dd6b4e0a081d4ed12e2af2")

  if err != nil {
    log.Fatal(err.Error())
  }

  fmt.Printf("Role found: %v\n", role)

  r := mux.NewRouter()
  r.HandleFunc("/", testCam.test)
  r.HandleFunc("/admin", adminAddCamera)
  r.HandleFunc("/closed", testCam.resolveClosed)
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer)))
  log.Fatal(http.ListenAndServe(":8080", r))
}

