package main

import (
	"log"
  "fmt"
  "os/exec"
  "runtime"
	"net/http"

	"github.com/gorilla/mux"
)

type Stream struct {
  command string
  options string
  framerate uint8
  //rtspTransport string
  bitrate string
  hlsTime uint8
  hlsWrap uint8
  codec string
  done chan bool
  err error
}

func (s *Stream) initiate(inputAddress string, outputFolder string) {
  s.command = ""
  s.done = make(chan bool)
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err == nil {
    s.command += string(path[0:len(path)-1])
    fmt.Printf("Path found: %s\n", string(path[:]))
  } else {
    log.Fatal(fmt.Sprintf("Could not find ffmpeg binary/executable! Error: %s", err.Error()))
  }

  // Put in options:
  if s.framerate == 0 {
    s.framerate = 30
  }
  if s.bitrate == "" {
    s.bitrate = "256K"
  }
  if s.hlsTime == 0 {
    s.hlsTime = 3
  }
  if s.hlsWrap == 0 {
    s.hlsWrap = 10
  }
  if s.codec == "" {
    s.codec = "copy"
  }

  cmd := exec.Command(s.command, "-r", fmt.Sprintf("%d", s.framerate), "-i", inputAddress, "-b:v", s.bitrate, "-hls_time", fmt.Sprintf("%d", s.hlsTime), "-hls_wrap", fmt.Sprintf("%d", s.hlsWrap), "-codec", s.codec, fmt.Sprintf("%s/stream.m3u8", outputFolder))
  fmt.Println(cmd.String())
  go func() {
    cmd.Run()
    s.done <- true
  }()
}

func (s *Stream) test(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  s.initiate("rtsp://170.93.143.139/rtplive/470011e600ef003a004ee33696235daa", "testStream")
  w.Write([]byte(`{"message": "Stream started"}`))
}

func (s *Stream) resolveClosed(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(fmt.Sprintf(`{"message": "%v"}`, <-s.done)))
}

type StreamServer struct {
}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")
  http.FileServer(http.Dir("./testStream/")).ServeHTTP(w, r)
}

func main() {
  stream1 := new(Stream)

  r := mux.NewRouter()
  r.HandleFunc("/", stream1.test)
  r.HandleFunc("/closed", stream1.resolveClosed)
  r.PathPrefix("/stream/").Handler(http.StripPrefix("/stream/", new(StreamServer)))
  log.Fatal(http.ListenAndServe(":8080", r))
}

