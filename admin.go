package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func adminCreateCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	hash := sha256.New()

	query := r.URL.Query()

	var session string
	var sid string
	var address string
	var room string
	var err error

	logger.Println(r.URL)
	if query["session"] == nil || query["sid"] == nil || query["address"] == nil || query["room"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	address = strings.ReplaceAll(query["address"][0], "\"", "\\\"") // Escape any double quotes for the command executer
	room = strings.ReplaceAll(query["room"][0], "\"", "\\\"")
	hash.Write([]byte(fmt.Sprintf("%s%s%d", address, room, time.Now().Unix())))
	id := fmt.Sprintf("%x", hash.Sum(nil))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Invalid parameters"}`))
		return
	}

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminCreateCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	//row, errRow := db.Query("SELECT * FROM cameras WHERE address=? OR room=?;", address, room)
	row, errRow := db.Query("SELECT * FROM cameras WHERE room=?;", room)
	if errRow != nil {
		logger.Printf("Error in adminCreateCamera checking for duplicated camera! Error: %s\n", err.Error())
	}
	if row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera with address or room code already created"}`)))
		return
	}
	_, err = db.Exec("INSERT INTO cameras VALUES ( ?, ?, ?, ?, 0, 0 );", sid, id, address, room)

	if err != nil {
		logger.Printf("Error in adminCreateCamera attempting to insert camera into database! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error creating camera record"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminReadCameras(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminReadCameras trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT id, address, room, lastStreamed, locked FROM cameras WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadCameras querying database for cameras! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get cameras"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			id           string
			address      string
			room         string
			lastStreamed uint64
			locked       uint64
		)

		if err := rows.Scan(&id, &address, &room, &lastStreamed, &locked); err != nil {
			logger.Printf("Error in adminReadCameras trying to scan row for camera values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for camera values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"id": "%s", "address": "%s", "room": %s, "lastStreamed": %d, "locked": %d}`, id, address, room, lastStreamed, locked)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "cameras": %s }`, jsonAccumulator)))
	return
}

func adminUpdateCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string
	var address string
	var room string
	var err error

	if query["id"] == nil || query["sid"] == nil || query["session"] == nil || query["address"] == nil || query["room"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]
	address = strings.ReplaceAll(query["address"][0], "\"", "\\\"")
	room = strings.ReplaceAll(query["room"][0], "\"", "\\\"")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Invalid parameters"}`))
		return
	}

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminUpdateCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM cameras WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminUpdateCamera trying to query database for requested camera! Error: %s\n", err.Error())
	}
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera to update does not exist!"}`)))
		return
	}
	_, err = db.Exec("UPDATE cameras SET address=?, room=? WHERE sid=? AND id=?;", address, room, sid, id)

	if err != nil {
		logger.Printf("Error in adminUpdateCamera attempting to execute update query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error updating camera record"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminDeleteCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string
	var id string

	if query["session"] == nil || query["sid"] == nil || query["id"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	id = query["id"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDeleteCamera checking session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	row, err := db.Query("SELECT * FROM cameras WHERE sid=? AND id=?;", sid, id)
	if err != nil {
		logger.Printf("Error in adminDeleteCamera trying to query database for requested camera! Error: %s\n", err.Error())
	}
	if !row.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera does not exist"}`)))
		return
	}
	_, err = db.Exec("DELETE FROM cameras WHERE sid=? AND id=?;", sid, id)

	if err != nil {
		logger.Printf("Error in adminDeleteCamera trying to execute delete query! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error deleting camera record"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))

	return
}

func adminStartCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session  string
		sid      string
		cameraId string
	)

	if query["sid"] == nil || query["session"] == nil || query["cameraId"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	cameraId = query["cameraId"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminStartCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT schools.address, cameras.address FROM cameras INNER JOIN schools ON cameras.sid=schools.id WHERE schools.id=? AND cameras.id=?;", sid, cameraId)

	if err != nil {
		logger.Printf("Error in adminStartCamera trying to query needed data for the camera! Error: %s\n", err.Error())
	}
	defer rows.Close()
	if !rows.Next() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Couldn't find camera id in database!"}`))
		return
	}

	var (
		schoolAddress string
		address       string
	)

	err = rows.Scan(&schoolAddress, &address)

	if err != nil {
		logger.Printf("Error in adminStartCamera trying to scan rows for data! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not get values from database"}`))
		return
	}

	client := new(http.Client)
	response, err := client.Get(fmt.Sprintf("%s/add/?id=%s&address=%s", schoolAddress, cameraId, address))

	if err != nil {
		logger.Printf("Error in adminStartCamera trying to request that the remote server starts the camera! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error starting camera feed"}`))
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	logger.Printf("Response received while starting camera: %s\n", string(body))
	if err != nil {
		logger.Printf("Error in adminStartCamera reading body from the request! Error: %s\n", err.Error())
	}
	if strings.Split(string(body), ";")[0] != "true" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Did not get response from camera api"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminStartAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session string
		sid     string
	)

	if query["sid"] == nil || query["session"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminStartAll trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT schools.address, cameras.address, cameras.id FROM cameras INNER JOIN schools ON cameras.sid=schools.id WHERE schools.id=? AND cameras.locked=0;", sid)

	if err != nil {
		logger.Printf("Error in adminStartAll trying to query needed data for the camera! Error: %s\n", err.Error())
	}

	updated := false

	defer rows.Close()

	for rows.Next() {
		updated = true

		var (
			schoolAddress string
			address       string
			cameraId      string
		)

		err = rows.Scan(&schoolAddress, &address, &cameraId)

		if err != nil {
			logger.Printf("Error in adminStartAll trying to scan rows for data! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Could not get values from database"}`))
			return
		}

		client := new(http.Client)
		response, err := client.Get(fmt.Sprintf("%s/add/?id=%s&address=%s", schoolAddress, cameraId, address))

		if err != nil {
			logger.Printf("Error in adminStartAll trying to request that the remote server starts the camera! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error starting camera feed"}`))
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		logger.Printf("Response received while starting camera: %s\n", string(body))
		if err != nil {
			logger.Printf("Error in adminStartAll reading body from the request! Error: %s\n", err.Error())
		}
		if strings.Split(string(body), ";")[0] != "true" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Did not get response from camera api"}`))
			return
		}
	}

	if !updated {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "No rows in database for cameras!"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminStopCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session  string
		sid      string
		cameraId string
	)

	if query["session"] == nil || query["sid"] == nil || query["cameraId"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	cameraId = query["cameraId"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminStopCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT schools.address FROM cameras INNER JOIN schools ON cameras.sid=schools.id WHERE schools.id=? AND cameras.id=?;", sid, cameraId)

	if err != nil {
		logger.Printf("Error in adminStopCamera trying to query needed data for the camera! Error: %s\n", err.Error())
	}
	defer rows.Close()
	if !rows.Next() {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Couldn't find camera id in database!"}`))
		return
	}

	var schoolAddress string

	err = rows.Scan(&schoolAddress)

	if err != nil {
		logger.Printf("Error in adminStopCamera trying to scan rows for data! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Could not get values from database"}`))
		return
	}

	client := new(http.Client)
	response, err := client.Get(fmt.Sprintf("%s/stop/?id=%s", schoolAddress, cameraId))

	if err != nil {
		logger.Printf("Error in adminStopCamera trying to request that the remote server stops the camera! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error stopping camera feed"}`))
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	logger.Printf("Response received while stopping camera: %s\n", string(body))
	if err != nil {
		logger.Printf("Error in adminStop reading body from the request! Error: %s\n", err.Error())
	}
	if strings.Split(string(body), ";")[0] != "true" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Did not get response from camera api"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": "Camera stopped"}`))
}

func adminStopAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session string
		sid     string
	)

	if query["sid"] == nil || query["session"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminStopAll trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT schools.address, cameras.id FROM cameras INNER JOIN schools ON cameras.sid=schools.id WHERE schools.id=? AND cameras.locked=0;", sid)

	if err != nil {
		logger.Printf("Error in adminStopAll trying to query needed data for the camera! Error: %s\n", err.Error())
	}

	updated := false

	defer rows.Close()

	for rows.Next() {
		updated = true

		var (
			schoolAddress string
			cameraId      string
		)

		err = rows.Scan(&schoolAddress, &cameraId)

		if err != nil {
			logger.Printf("Error in adminStopAll trying to scan rows for data! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Could not get values from database"}`))
			return
		}

		client := new(http.Client)
		response, err := client.Get(fmt.Sprintf("%s/stop/?id=%s", schoolAddress, cameraId))

		if err != nil {
			logger.Printf("Error in adminStopAll trying to request that the remote server stops the camera! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Error stopping camera feed"}`))
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		logger.Printf("Response received while stopping camera: %s\n", string(body))
		if err != nil {
			logger.Printf("Error in adminStopAll reading body from the request! Error: %s\n", err.Error())
		}
		if strings.Split(string(body), ";")[0] != "true" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status": false, "err": "Did not get response from camera api"}`))
			return
		}
	}

	if !updated {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "No rows in database for cameras!"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminLockCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session  string
		sid      string
		cameraId string
	)

	if query["sid"] == nil || query["session"] == nil || query["cameraId"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	cameraId = query["cameraId"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminLockCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	_, err := db.Query("UPDATE cameras SET locked=1 WHERE sid=? AND id=?;", sid, cameraId)

	if err != nil {
		logger.Printf("Error in adminLockCamera trying to update locked status for the camera! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error trying to lock camera!"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminUnlockCamera(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var (
		session  string
		sid      string
		cameraId string
	)

	if query["sid"] == nil || query["session"] == nil || query["cameraId"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]
	cameraId = query["cameraId"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminUnlockCamera trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	_, err := db.Query("UPDATE cameras SET locked=0 WHERE sid=? AND id=?;", sid, cameraId)

	if err != nil {
		logger.Printf("Error in adminUnlockCamera trying to unlock the camera! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": false, "err": "Error trying to unlock camera!"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": true, "err": ""}`))
}

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	var session string
	var sid string

	if query["session"] == nil || query["sid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
		return
	}

	session = query["session"][0]
	sid = query["sid"][0]

	if role, err := checkSession(sid, session); role != "A" {
		if err != nil {
			logger.Printf("Error in adminDashboard trying to check session! Error: %s\n", err.Error())
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
		return
	}

	rows, err := db.Query("SELECT id, address, room, lastStreamed, locked FROM cameras WHERE sid=?;", sid)
	if err != nil {
		logger.Printf("Error in adminReadCameras querying database for cameras! Error: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get cameras"}`)))
		return
	}
	defer rows.Close()

	jsonAccumulator := "["

	for rows.Next() {
		var (
			id           string
			address      string
			room         string
			lastStreamed uint64
			locked       uint64
		)

		if err := rows.Scan(&id, &address, &room, &lastStreamed, &locked); err != nil {
			logger.Printf("Error in adminReadCameras trying to scan row for camera values! Error: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for camera values"}`)))
			return
		}

		if jsonAccumulator != "[" {
			jsonAccumulator += ","
		}

		jsonAccumulator += fmt.Sprintf(`{"id": "%s", "address": "%s", "room": %s, "lastStreamed": %d, "locked": %d}`, id, address, room, lastStreamed, locked)
	}

	jsonAccumulator += "]"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status": true, "err": false, "cameras": %s }`, jsonAccumulator)))
	return
}
