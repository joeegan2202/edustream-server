package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

var (
  cameras []*Camera
)

type Camera struct {
  id string
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
  c.streamCommand = ""
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  syscall.Umask(0)
  os.MkdirAll(fmt.Sprintf("streams/%s", c.outputFolder), 0755)

  if err != nil {
    return fmt.Errorf("Could not find ffmpeg binary/executable! Error: %s", err.Error())
  }

  c.streamCommand += string(path[0:len(path)-1])
  fmt.Printf("Path found: %s\n", c.streamCommand)

  c.streamCmd = exec.Command(c.streamCommand, "-i", c.inputAddress, "-hls_time", fmt.Sprintf("%d", c.streamHlsTime), "-hls_wrap", fmt.Sprintf("%d", c.streamHlsWrap), "-codec", "copy", fmt.Sprintf("streams/%s/stream.m3u8", c.outputFolder))
  fmt.Println(c.streamCmd.String())
  go func() {
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
  c.recordCommand = ""
  // Change with OS:
  var path []byte
  var err error
  if runtime.GOOS == "windows" {
    path, err = exec.Command("where ffmpeg").Output()
  } else {
    path, err = exec.Command("/usr/bin/which", "ffmpeg").Output()
  }

  syscall.Umask(0)
  os.MkdirAll(fmt.Sprintf("streams/%s", c.outputFolder), 0755)

  if err != nil {
    return fmt.Errorf("Could not find ffmpeg binary/executable! Error: %s", err.Error())
  }

  c.recordCommand += string(path[0:len(path)-1])
  fmt.Printf("Path found: %s\n", c.recordCommand)

  c.recordCmd = exec.Command(c.recordCommand, "-i", c.inputAddress, "-y", "-codec", "copy", fmt.Sprintf("streams/%s/record.mp4", c.outputFolder))
  fmt.Println(c.recordCmd.String())
  go func() {
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

