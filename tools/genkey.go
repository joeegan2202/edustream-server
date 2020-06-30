package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"time"

	"log"
  "fmt"
)

func generateKey() {
  key, err := rsa.GenerateKey(rand.Reader, 2048)

  if err != nil {
    log.Fatalf("Error trying to generate key! %s\n", err.Error())
  }

  now := time.Now().Format(time.RFC3339)
  file, err := os.OpenFile(fmt.Sprintf("keygen-%s.pem", now), os.O_RDWR|os.O_CREATE, 0755)

  if err != nil {
    log.Fatalf("Error trying to open keyfile! %s\n", err.Error())
  }

  filepub, err := os.OpenFile(fmt.Sprintf("keygen-%s.pub", now), os.O_RDWR|os.O_CREATE, 0755)

  if err != nil {
    log.Fatalf("Error trying to open keyfile! %s\n", err.Error())
  }

  _, err = file.Write(x509.MarshalPKCS1PrivateKey(key))

  if err != nil {
    log.Fatalf("Error trying to write to keyfile! %s\n", err.Error())
  }

  filepub.Write([]byte(fmt.Sprintf("%x", x509.MarshalPKCS1PublicKey(&key.PublicKey))))
  fmt.Printf("Keyfile has been generated! Public key: %x\n", x509.MarshalPKCS1PublicKey(&key.PublicKey))
}
