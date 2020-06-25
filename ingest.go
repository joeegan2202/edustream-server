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

  dir := r.URL.EscapedPath()[:strings.LastIndex(r.URL.EscapedPath(), "/")]

  fmt.Println(dir)

  os.MkdirAll(fmt.Sprintf("%s/%s", os.Getenv("FS_PATH"), dir), 0755)
  file, err := os.OpenFile(fmt.Sprintf("%s/%s", os.Getenv("FS_PATH"), r.URL.EscapedPath()), os.O_RDWR|os.O_CREATE, 0755)

  if err != nil {
    log.Fatal(err.Error())
  }

  io.Copy(file, r.Body)
}

