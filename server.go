package main

import (
	"fmt"
	"log"
  "os"
	"net/http"
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

  db = loadDatabase()

  createTables(db)

  r := mux.NewRouter()
  r.HandleFunc("/admin/start/camera/", adminStartCamera) // Admins can start and stop cameras
  r.HandleFunc("/admin/start/all/", adminStartAll) // Admins can start and stop all of the available cameras
  r.HandleFunc("/admin/stop/camera/", adminStopCamera)
  r.HandleFunc("/admin/stop/all/", adminStopAll) // Admins can start and stop all of the available cameras
  r.HandleFunc("/admin/create/camera/", adminCreateCamera) // CRUD operations for camera management
  r.HandleFunc("/admin/read/camera/", adminReadCameras)
  r.HandleFunc("/admin/update/camera/", adminUpdateCamera)
  r.HandleFunc("/admin/delete/camera/", adminDeleteCamera)
  r.HandleFunc("/admin/import/people/", adminImportPeople)
  r.HandleFunc("/admin/import/classes/", adminImportClasses)
  r.HandleFunc("/admin/import/roster/", adminImportRoster)
  r.HandleFunc("/admin/import/periods/", adminImportPeriods)
  r.HandleFunc("/admin/read/people/", adminReadPeople)
  r.HandleFunc("/admin/read/classes/", adminReadClasses)
  r.HandleFunc("/admin/read/roster/", adminReadRoster)
  r.HandleFunc("/admin/read/periods/", adminReadPeriods)
  r.HandleFunc("/auth/", tempAuthorize)
  r.HandleFunc("/status/", receiveStatus)
  //r.HandleFunc("/request/", requestStream) // For admins/teachers/students who are requesting a video stream
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
