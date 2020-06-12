package main

import (
  "os/exec"
  "runtime"
  "fmt"
  "log"
)

type Stream struct {
  command string
  framerate uint8
  bitrate string
  hlsTime uint8
  hlsWrap uint8
  codec string
  done chan bool
  err error
}

type Record struct {
  command string
  framerate uint8
  bitrate string
  codec string
  done chan bool
  err error
}

type Camera struct {
  stream *Stream
  record *Record
}

func (c *Camera) initiate(inputAddress string, outputFolder string) {
  c.stream.initiate(inputAddress, outputFolder)
  c.record.initiate(inputAddress, outputFolder)
}

func (r *Record) initiate(inputAddress string, outputFolder string) {
  r.command = ""
  r.done = make(chan bool)
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err == nil {
    r.command += string(path[0:len(path)-1])
    fmt.Printf("Path found: %s\n", string(path[:]))
  } else {
    log.Fatal(fmt.Sprintf("Could not find ffmpeg binary/executable! Error: %s", err.Error()))
  }

  // Put in options:
  if r.framerate == 0 {
    r.framerate = 30
  }
  if r.bitrate == "" {
    r.bitrate = "256K"
  }
  if r.codec == "" {
    r.codec = "copy"
  }

  cmd := exec.Command(r.command, "-r", fmt.Sprintf("%d", r.framerate), "-i", inputAddress, "-b:v", r.bitrate, "-codec", r.codec, "-y", fmt.Sprintf("%s/record.mp4", outputFolder))
  fmt.Println(cmd.String())
  go func() {
    cmd.Run()
    r.done <- true
  }()
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

