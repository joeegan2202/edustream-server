package main

import (
	"crypto/sha256"
	"fmt"
  "os"
  "log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
  "syscall"
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
  var hlsTime uint64
  var hlsWrap uint64
  var err error

  logger.Println(r.URL)
  if query["session"] == nil || query["sid"] == nil || query["address"] == nil || query["room"] == nil || query["hlsTime"] == nil || query["hlsWrap"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  sid = query["sid"][0]
  address = strings.ReplaceAll(query["address"][0], "\"", "\\\"") // Escape any double quotes for the command executer
  room = strings.ReplaceAll(query["room"][0], "\"", "\\\"")
  hlsTime, err = strconv.ParseUint(query["hlsTime"][0], 10, 64)
  hlsWrap, err = strconv.ParseUint(query["hlsWrap"][0], 10, 64)
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

  row, errRow := db.Query("SELECT * FROM cameras WHERE address=? OR room=?;", address, room)
  if errRow != nil {
    logger.Printf("Error in adminCreateCamera checking for duplicated camera! Error: %s\n", err.Error())
  }
  if row.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera with address or room code already created"}`)))
    return
  }
  _, err = db.Exec("INSERT INTO cameras VALUES ( ?, ?, ?, ?, ?, ? );", sid, id, address, room, hlsTime, hlsWrap)

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

  rows, err := db.Query("SELECT id, address, room, hlsTime, hlsWrap FROM cameras WHERE sid=?;", sid)
	if err != nil {
    logger.Printf("Error in adminCreateCamera querying database for cameras! Error: %s\n", err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to get cameras"}`)))
    return
	}
	defer rows.Close()

  jsonAccumulator := "["

  for rows.Next() {
		var (
			id string
			address string
      room string
      hlsTime uint64
      hlsWrap uint64
      streaming bool
      recording bool
		)

		if err := rows.Scan(&id, &address, &room, &hlsTime, &hlsWrap); err != nil {
      logger.Printf("Error in adminReadCameras trying to scan row for camera values! Error: %s\n", err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Failed to scan rows for camera values"}`)))
      return
		}

    if jsonAccumulator != "[" {
      jsonAccumulator += ","
    }

    streaming = false
    recording = false

    for _, c := range cameras {
      if c.id == id {
        streaming = true
        recording = true
        break
      }
    }

    jsonAccumulator += fmt.Sprintf(`{"id": "%s", "address": "%s", "room": %s, "hlsTime": %d, "hlsWrap": %d, "streaming": %t, "recording": %t}`, id, address, room, hlsTime, hlsWrap, streaming, recording)
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
  var hlsTime uint64
  var hlsWrap uint64
  var err error

  if query["id"] == nil || query["sid"] == nil || query["session"] == nil || query["address"] == nil || query["room"] == nil || query["hlsTime"] == nil || query["hlsWrap"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  sid = query["sid"][0]
  id = query["id"][0]
  address = strings.ReplaceAll(query["address"][0], "\"", "\\\"")
  room = strings.ReplaceAll(query["room"][0], "\"", "\\\"")
  hlsTime, err = strconv.ParseUint(query["hlsTime"][0], 10, 64)
  hlsWrap, err = strconv.ParseUint(query["hlsWrap"][0], 10, 64)

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
  if err != nil  {
    logger.Printf("Error in adminUpdateCamera trying to query database for requested camera! Error: %s\n", err.Error())
  }
  if !row.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera to update does not exist!"}`)))
    return
  }
  _, err = db.Exec("UPDATE cameras SET address=?, room=?, hlsTime=?, hlsWrap=?, WHERE sid=? AND id=?;", address, room, hlsTime, hlsWrap, sid, id)

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
    session string
    sid string
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

  for _, c := range cameras {
    if c.id == cameraId {
      w.WriteHeader(http.StatusBadRequest)
      w.Write([]byte(`{"status": false, "err": "Camera already started"}`))
      return
    }
  }

  rows, err := db.Query("SELECT schools.address, cameras.address, cameras.room, cameras.hlsTime, cameras.hlsWrap FROM cameras INNER JOIN schools ON cameras.sid=schools.id WHERE schools.id=? AND cameras.id=?;", sid, cameraId)

  if err != nil {
    logger.Printf("Error in adminStartCamera trying to query needed data for the camera! Error: %s\n", err.Error())
  }
  if !rows.Next() {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Couldn't find camera id in database!"}`))
    return
  }
  defer rows.Close()

  camera := new(Camera)

  var (
    schoolAddress string
    address string
    room string
    hlsTime uint64
    hlsWrap uint64
  )

  err = rows.Scan(&schoolAddress, &address, &room, &hlsTime, &hlsWrap)

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
    w.Write([]byte(`{"status": false, "err": "Error starting camera api"}`))
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

  camera.inputAddress = fmt.Sprintf("%s/stream/%s/stream.m3u8", schoolAddress, cameraId)
  camera.outputFolder = fmt.Sprintf("%s/%s", sid, room)
  syscall.Umask(0)
  os.MkdirAll(fmt.Sprintf("streams/%s", camera.outputFolder), 0755)
  camera.id = cameraId
  camera.streamHlsTime = hlsTime
  camera.streamHlsWrap = hlsWrap

  f, err := os.OpenFile(fmt.Sprintf("streams/%s/logfile-%s.txt", camera.outputFolder, time.Now().Format(time.RFC3339)), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    logger.Printf("Error in adminStartCamera opening log file for camera! Error: %s\n", err.Error())
    return
  }
  defer f.Close()

  camera.logger = log.New(f, "", log.Ldate | log.Ltime)

  cameras = append(cameras, camera)

  err = camera.initiateStream()
  if err != nil {
    logger.Printf("Error in adminCreateCamera trying to initiate stream! Error: %s\n", err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Could not initiate the camera's stream"}`))
    return
  }

  err = camera.initiateRecord()
  if err != nil {
    logger.Printf("Error in adminCreateCamera trying to initiate record! Error: %s\n", err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Could not initiate the camera's record"}`))
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
    session string
    sid string
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

  for _, c := range cameras {
    if c.id == cameraId {
      c.streamCmd.Process.Kill()
      c.recordCmd.Process.Kill()
      w.WriteHeader(http.StatusOK)
      w.Write([]byte(`{"status": true, "err": ""}`))
      return
    }
  }

  w.WriteHeader(http.StatusBadRequest)
  w.Write([]byte(`{"status": false, "err": "Camera not started"}`))
  return
}
