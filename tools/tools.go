package main

import (
	"fmt"
	"os"
)

func main() {
	switch os.Args[1] {
	case "genkey", "-genkey", "--genkey":
		generateKey()
	case "teacher", "-teacher", "--teacher":
		teacherRoster()
	case "help", "-help", "--help":
		printHelp()
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
