package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var logger *log.Logger

func main() {
	f, err := os.OpenFile(fmt.Sprintf("logfile-%s.txt", time.Now().Format(time.RFC3339)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	logger = log.New(f, "", log.Ldate|log.Ltime)

	db = loadDatabase()

	createTables(db)

	go manageCameras()

	r := mux.NewRouter()
	r.HandleFunc("/admin/start/camera/", adminStartCamera) // Admins can start and stop cameras
	r.HandleFunc("/admin/start/all/", adminStartAll)       // Admins can start and stop all of the available cameras
	r.HandleFunc("/admin/stop/camera/", adminStopCamera)
	r.HandleFunc("/admin/stop/all/", adminStopAll)           // Admins can start and stop all of the available cameras
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

func manageCameras() {
	selectq, err := db.Prepare("SELECT schools.address, cameras.id, cameras.address FROM cameras INNER JOIN classes ON cameras.room=classes.room INNER JOIN periods ON periods.code=classes.period INNER JOIN schools ON schools.id=cameras.sid WHERE periods.stime<? AND periods.etime>?;")
	if err != nil {
		logger.Panicf("Couldn't initialize starting select statement! %s\n", err.Error())
	}
	selectw, err := db.Prepare("SELECT schools.address, cameras.id FROM cameras INNER JOIN classes ON cameras.room=classes.room INNER JOIN periods ON periods.code=classes.period INNER JOIN schools ON schools.id=cameras.sid WHERE (periods.stime>? OR periods.etime<?) AND cameras.streaming=1;")
	if err != nil {
		logger.Panicf("Couldn't initialize stopping select statement! %s\n", err.Error())
	}
	client := new(http.Client)

	for {
		wait := time.After(5 * time.Second)

		now := time.Now().Unix()
		rows, err := selectq.Query(now, now)

		if err != nil {
			logger.Printf("Error trying to query database to automatically start cameras! %s\n", err.Error())
			continue
		}

		for rows.Next() {
			var (
				schoolAddress string
				cameraId      string
				cameraAddress string
			)

			err := rows.Scan(&schoolAddress, &cameraId, &cameraAddress)

			if err != nil {
				logger.Printf("Error trying to scan rows to automatically start cameras! %s\n", err.Error())
				rows.Close()
				continue
			}

			response, err := client.Get(fmt.Sprintf("%s/add/?id=%s&address=%s", schoolAddress, cameraId, cameraAddress))

			if err != nil {
				logger.Printf("Error trying to request that the remote server starts the camera automatically! Error: %s\n", err.Error())
				rows.Close()
				continue
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				logger.Printf("Error reading body from the request to automatically start camera! Error: %s\n", err.Error())
			}
			if strings.Split(string(body), ";")[0] != "true" {
				logger.Printf("Remote server failed to start camera! Error: %s\n", strings.Split(string(body), ";")[1])
				rows.Close()
				continue
			}
		}

		rows, err = selectw.Query(now, now)

		if err != nil {
			logger.Printf("Error trying to query database to automatically stop cameras! %s\n", err.Error())
			continue
		}

		for rows.Next() {
			var (
				schoolAddress string
				cameraId      string
			)

			err := rows.Scan(&schoolAddress, &cameraId)

			if err != nil {
				logger.Printf("Error trying to scan rows to automatically stop cameras! %s\n", err.Error())
				rows.Close()
				continue
			}

			response, err := client.Get(fmt.Sprintf("%s/stop/?id=%s", schoolAddress, cameraId))

			if err != nil {
				logger.Printf("Error trying to request that the remote server stops the camera automatically! Error: %s\n", err.Error())
				rows.Close()
				continue
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				logger.Printf("Error reading body from the request to automatically stop camera! Error: %s\n", err.Error())
			}
			if strings.Split(string(body), ";")[0] != "true" {
				logger.Printf("Remote server failed to stop camera! Error: %s\n", strings.Split(string(body), ";")[1])
				rows.Close()
				continue
			}
		}

		rows.Close()
		<-wait
	}
}
