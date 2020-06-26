package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type IngestServer struct {}

func (i *IngestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  fmt.Println(r.URL.EscapedPath())

  cid := r.URL.EscapedPath()[:strings.LastIndex(r.URL.EscapedPath(), "/")]
  filename := r.URL.EscapedPath()[strings.LastIndex(r.URL.EscapedPath(), "/"):]

  rows, err := db.Query("SELECT sid, room FROM cameras WHERE id=?;", cid)

  if err != nil || !rows.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Could not get camera from database!"))
    return
  }

  defer rows.Close()

  var room string
  var sid string

  if err = rows.Scan(&sid, &room); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Failed to get room from query!"))
    return
  }

  dir := fmt.Sprintf("%s/%s/%s", os.Getenv("FS_PATH"), sid, room)

  fmt.Println(fmt.Sprintf("%s/%s", dir, filename))

  os.MkdirAll(dir, 0755)
  file, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, filename), os.O_RDWR|os.O_CREATE, 0755)

  if err != nil {
    log.Fatal(err.Error())
  }

  io.Copy(file, r.Body)
}

