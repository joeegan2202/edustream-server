package main

import (
  "os/exec"
  "runtime"
  "fmt"
  "log"
)

var (
  cameras []*Camera
)

type Camera struct {
  streamCommand string
  streamFramerate uint8
  streamBitrate string
  streamHlsTime uint8
  streamHlsWrap uint8
  streamCodec string
  streamDone chan bool
  streamErr error
  recordCommand string
  recordFramerate uint8
  recordBitrate string
  recordCodec string
  recordDone chan bool
  recordErr error
  inputAddress string
  outputFolder string
}

func (c *Camera) initiate() {
  c.initiateStream()
  c.initiateRecord()
}

func (c *Camera) initiateStream() {
  c.streamCommand = ""
  c.streamDone = make(chan bool)
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err == nil {
    c.streamCommand += string(path[0:len(path)-1])
    fmt.Printf("Path found: %s\n", string(path[:]))
  } else {
    log.Fatal(fmt.Sprintf("Could not find ffmpeg binary/executable! Error: %s", err.Error()))
  }

  // Put in options:
  if c.streamFramerate == 0 {
    c.streamFramerate = 30
  }
  if c.streamBitrate == "" {
    c.streamBitrate = "256K"
  }
  if c.streamHlsTime == 0 {
    c.streamHlsTime = 3
  }
  if c.streamHlsWrap == 0 {
    c.streamHlsWrap = 10
  }
  if c.streamCodec == "" {
    c.streamCodec = "copy"
  }

  cmd := exec.Command(c.streamCommand, "-r", fmt.Sprintf("%d", c.streamFramerate), "-i", c.inputAddress, "-b:v", c.streamBitrate, "-hls_time", fmt.Sprintf("%d", c.streamHlsTime), "-hls_wrap", fmt.Sprintf("%d", c.streamHlsWrap), "-codec", c.streamCodec, fmt.Sprintf("%s/stream.m3u8", c.outputFolder))
  fmt.Println(cmd.String())
  go func() {
    cmd.Run()
    c.streamDone <- true
  }()
}

func (c *Camera) initiateRecord() {
  c.recordCommand = ""
  c.recordDone = make(chan bool)
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err == nil {
    c.recordCommand += string(path[0:len(path)-1])
    fmt.Printf("Path found: %s\n", string(path[:]))
  } else {
    log.Fatal(fmt.Sprintf("Could not find ffmpeg binary/executable! Error: %s", err.Error()))
  }

  // Put in options:
  if c.recordFramerate == 0 {
    c.recordFramerate = 30
  }
  if c.recordBitrate == "" {
    c.recordBitrate = "256K"
  }
  if c.recordCodec == "" {
    c.recordCodec = "copy"
  }

  cmd := exec.Command(c.recordCommand, "-r", fmt.Sprintf("%d", c.recordFramerate), "-i", c.inputAddress, "-b:v", c.recordBitrate, "-codec", c.recordCodec, "-y", fmt.Sprintf("%s/record.mp4", c.outputFolder))
  fmt.Println(cmd.String())
  go func() {
    cmd.Run()
    c.recordDone <- true
  }()
}
