package main

import (
	"fmt"
  "log"
	"os/exec"
	"runtime"
)

var (
  cameras []*Camera
)

type Camera struct {
  id string
  logger *log.Logger
  streamCmd *exec.Cmd
  recordCmd *exec.Cmd
  streamCommand string
  streamHlsTime uint64
  streamHlsWrap uint64
  recordCommand string
  inputAddress string
  outputFolder string
}

func (c *Camera) initiateStream() error {
  c.logger.Println("Initiating stream")

  c.streamCommand = ""
  // Change with OS:
  var path []byte
  var err error

  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err != nil {
    c.logger.Printf("Could not find binary/executable! Error: %s", err.Error())
    return fmt.Errorf("Could not find ffmpeg binary/executable! Error: %s", err.Error())
  }

  c.streamCommand += string(path[0:len(path)-1])

  c.streamCmd = exec.Command(c.streamCommand, "-i", c.inputAddress, "-hls_time", fmt.Sprintf("%d", c.streamHlsTime), "-hls_wrap", fmt.Sprintf("%d", c.streamHlsWrap), "-codec", "copy", fmt.Sprintf("streams/%s/stream.m3u8", c.outputFolder))
  logger.Printf("Stream starting with command: %s\n", c.streamCmd.String())
  go func() {
    c.streamCmd.Stderr = c.logger.Writer()
    c.streamCmd.Run()
    index := -1
    for i, camera := range cameras {
      if camera.id == c.id {
        index = i
        break
      }
    }
    if index == -1 {
      return
    }
    cameras[len(cameras)-1], cameras[index] = cameras[index], cameras[len(cameras)-1]
    cameras = cameras[:len(cameras)-1] // Magic code to delete this camera from the list of cameras
  }()

  return nil
}

func (c *Camera) initiateRecord() error {
  c.logger.Println("Initiating recording")

  c.recordCommand = ""
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  if err != nil {
    c.logger.Printf("Could not get binary! Error: %s\n", err.Error())
    return fmt.Errorf("Could not find ffmpeg binary/executable! Error: %s", err.Error())
  }

  c.recordCommand += string(path[0:len(path)-1])

  c.recordCmd = exec.Command(c.recordCommand, "-i", c.inputAddress, "-y", "-codec", "copy", fmt.Sprintf("streams/%s/record.mp4", c.outputFolder))
  logger.Printf("Record starting with command: %s\n", c.recordCmd.String())
  go func() {
    c.recordCmd.Stderr = c.logger.Writer()
    c.recordCmd.Run()
    index := -1
    for i, camera := range cameras {
      if camera.id == c.id {
        index = i
        break
      }
    }
    if index == -1 {
      return
    }
    cameras[len(cameras)-1], cameras[index] = cameras[index], cameras[len(cameras)-1]
    cameras = cameras[:len(cameras)-1] // Magic code to delete this camera from the list of cameras
  }()

  return nil
}

