package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func adminCreateCamera(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Content-Type", "application/json")

  hash := sha256.New()

  query := r.URL.Query()

  var session string
  var address string
  var room string
  var framerate uint64
  var bitrate string
  var hlsTime uint64
  var hlsWrap uint64
  var codec string
  var err error

  if query["session"] == nil || query["address"] == nil || query["room"] == nil || query["framerate"] == nil || query["bitrate"] == nil || query["hlsTime"] == nil || query["hlsWrap"] == nil || query["codec"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  address = query["address"][0]
  room = query["room"][0]
  framerate, err = strconv.ParseUint(query["framerate"][0], 10, 64)
  bitrate = query["bitrate"][0]
  hlsTime, err = strconv.ParseUint(query["hlsTime"][0], 10, 64)
  hlsWrap, err = strconv.ParseUint(query["hlsWrap"][0], 10, 64)
  codec = query["codec"][0]
  hash.Write([]byte(fmt.Sprintf("%s%s%d", address, room, time.Now().Unix())))
  id := fmt.Sprintf("%x", hash.Sum(nil))

  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Invalid parameters"}`))
    return
  }

  if role, _ := checkSession(session); role != "A" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
    return
  }

  row, errRow := db.Query("SELECT * FROM cameras WHERE address=? OR room=?;", address, room)
  if errRow != nil || row.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera with address or room code already created"}`)))
    return
  }
  _, err = db.Exec("INSERT INTO cameras VALUES (?, ?, ?, ?, ?, ?, ?, ?);", id, address, room, framerate, bitrate, hlsTime, hlsWrap, codec)

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

func adminReadCameras(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  w.Header().Set("Content-Type", "application/json")

  query := r.URL.Query()

  var session string

  if query["session"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing session"}`))
    return
  }

  session = query["session"][0]

  if role, _ := checkSession(session); role != "A" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
    return
  }

  rows, err := db.Query("SELECT * FROM cameras;")
	if err != nil {
		fmt.Println(err.Error())
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
      framerate uint64
      bitrate string
      hlsTime uint64
      hlsWrap uint64
      codec string
      streaming bool
      recording bool
		)

		if err := rows.Scan(&id, &address, &room, &framerate, &bitrate, &hlsTime, &hlsWrap, &codec); err != nil {
      fmt.Println(err.Error())
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

    jsonAccumulator += fmt.Sprintf(`{"id": "%s", "address": "%s", "room": %s, "framerate": %d, "bitrate": "%s", "hlsTime": %d, "hlsWrap": %d, "codec": "%s", "streaming": %t, "recording": %t}`, id, address, room, framerate, bitrate, hlsTime, hlsWrap, codec, streaming, recording)
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
  var id string
  var address string
  var room string
  var framerate uint64
  var bitrate string
  var hlsTime uint64
  var hlsWrap uint64
  var codec string
  var err error

  if query["id"] == nil || query["session"] == nil || query["address"] == nil || query["room"] == nil || query["framerate"] == nil || query["bitrate"] == nil || query["hlsTime"] == nil || query["hlsWrap"] == nil || query["codec"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  id = query["id"][0]
  address = query["address"][0]
  room = query["room"][0]
  framerate, err = strconv.ParseUint(query["framerate"][0], 10, 64)
  bitrate = query["bitrate"][0]
  hlsTime, err = strconv.ParseUint(query["hlsTime"][0], 10, 64)
  hlsWrap, err = strconv.ParseUint(query["hlsWrap"][0], 10, 64)
  codec = query["codec"][0]

  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Invalid parameters"}`))
    return
  }

  if role, _ := checkSession(session); role != "A" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
    return
  }

  row, errRow := db.Query("SELECT * FROM cameras WHERE id=?;", id)
  if errRow != nil || !row.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera to update does not exist!"}`)))
    return
  }
  _, err = db.Exec("UPDATE cameras SET address=?, room=?, framerate=?, bitrate=?, hlsTime=?, hlsWrap=?, codec=? WHERE id=?;", address, room, framerate, bitrate, hlsTime, hlsWrap, codec, id)

  if err != nil {
    fmt.Printf("Error found: %s\n", err.Error())
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
  var id string

  if query["session"] == nil || query["id"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  id = query["id"][0]

  if role, _ := checkSession(session); role != "A" {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Incorrect role for session"}`)))
    return
  }

  row, errRow := db.Query("SELECT * FROM cameras WHERE id=?;", id)
  if errRow != nil || !row.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf(`{"status": false, "err": "Camera does not exist"}`)))
    return
  }
  _, err := db.Exec("DELETE FROM cameras WHERE id=?;", id)

  if err != nil {
    fmt.Printf("Error found: %s\n", err.Error())
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
    cameraId string
  )

  if query["session"] == nil || query["cameraId"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  cameraId = query["cameraId"][0]

  if role, _ := checkSession(session); role != "A" {
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

  rows, err := db.Query("SELECT * FROM cameras WHERE id=?;", cameraId)

  if err != nil || !rows.Next() {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Couldn't find camera id in database!"}`))
    return
  }
  defer rows.Close()

  camera := new(Camera)

  var (
    id string
    address string
    room string
    framerate uint64
    bitrate string
    hlsTime uint64
    hlsWrap uint64
    codec string
  )

  err = rows.Scan(&id, &address, &room, &framerate, &bitrate, &hlsTime, &hlsWrap, &codec)

  if err != nil {
    fmt.Println(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Could not get values from database"}`))
    return
  }

  camera.inputAddress = address
  camera.outputFolder = room
  camera.id = id
  camera.streamBitrate = bitrate
  camera.recordBitrate = bitrate
  camera.streamCodec = codec
  camera.recordCodec = codec
  camera.streamFramerate = framerate
  camera.recordFramerate = framerate
  camera.streamHlsTime = hlsTime
  camera.streamHlsWrap = hlsWrap

  cameras = append(cameras, camera)

  err = camera.initiateStream()
  if err != nil {
    fmt.Println(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"status": false, "err": "Could not initiate the camera's stream"}`))
    return
  }

  err = camera.initiateRecord()
  if err != nil {
    fmt.Println(err.Error())
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
    cameraId string
  )

  if query["session"] == nil || query["cameraId"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"status": false, "err": "Missing parameters"}`))
    return
  }

  session = query["session"][0]
  cameraId = query["cameraId"][0]

  if role, _ := checkSession(session); role != "A" {
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
