package main

import (
	"fmt"
	"os"
)

func main() {
  switch os.Args[1] {
  case "genkey":
  case "-genkey":
  case "--genkey":
    generateKey()
    break
  case "help":
  case "-help":
  case "--help":
    printHelp()
    break
  }
}

func printHelp() {
  fmt.Println(`EduStream Management Tools:
  Usage: tools (tool) [options]
  Tools:
  --help - Prints this help page
  --genkey - Generates a new key to send to a deploy server, and prints the public key
  `)
}
