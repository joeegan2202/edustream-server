package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type IngestServer struct {}

func (i *IngestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  query := r.URL.Query()

  if query["signature"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Incorrect parameters! Missing signature"))
    return
  }

  signature := query["signature"][0]

  fmt.Println(r.URL.EscapedPath())

  cid := r.URL.EscapedPath()[:strings.LastIndex(r.URL.EscapedPath(), "/")]
  filename := r.URL.EscapedPath()[strings.LastIndex(r.URL.EscapedPath(), "/"):]

  rows, err := db.Query("SELECT cameras.sid, cameras.room, schools.publicKey FROM cameras INNER JOIN schools ON schools.id=cameras.sid WHERE cameras.id=?;", cid)

  if err != nil || !rows.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Could not get camera from database!"))
    return
  }

  defer rows.Close()

  var (
    room string
    sid string
    publicKeyString string
    publicKey *rsa.PublicKey
  )

  if err = rows.Scan(&sid, &room, &publicKeyString); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Failed to get room from query!"))
    return
  }

  publicKeyBytes, err := hex.DecodeString(publicKeyString)
  publicKey, err = x509.ParsePKCS1PublicKey(publicKeyBytes)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Could not parse public key from database!"))
    return
  }

  signData := make([]byte, 2048)

  io.ReadAtLeast(r.Body, signData, 100)

  hasher := sha256.New()

  hasher.Write(signData)

  signBytes, err := hex.DecodeString(signature)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Cannot decode hex data from signature parameter!"))
    return
  }

  err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), signBytes)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Signature did not verify!"))
    return
  }

  dir := fmt.Sprintf("%s/%s/%s", os.Getenv("FS_PATH"), sid, room)

  fmt.Println(fmt.Sprintf("%s/%s", dir, filename))

  os.MkdirAll(dir, 0755)
  file, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, filename), os.O_RDWR|os.O_CREATE, 0755)

  if err != nil {
    log.Fatal(err.Error())
  }

  io.Copy(file, io.MultiReader(bytes.NewReader(signData), r.Body))
}

func receiveStatus(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Access-Control-Allow-Origin", "*")

  query := r.URL.Query()

  if query["signature"] == nil || query["cameraId"] == nil || query["status"] == nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Incorrect parameters! Missing parameters"))
    return
  }

  signature := query["signature"][0]
  cid := query["cameraId"][0]
  status, err := strconv.Atoi(query["status"][0])

  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Invalid value found for status!"))
    return
  }

  rows, err := db.Query("SELECT schools.publicKey FROM cameras INNER JOIN schools ON schools.id=cameras.sid WHERE cameras.id=?;", cid)

  if err != nil || !rows.Next() {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte("Could not get camera from database!"))
    return
  }

  defer rows.Close()

  var (
    publicKeyString string
    publicKey *rsa.PublicKey
  )

  if err = rows.Scan(&publicKeyString); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Failed to get room from query!"))
    return
  }

  publicKeyBytes, err := hex.DecodeString(publicKeyString)
  publicKey, err = x509.ParsePKCS1PublicKey(publicKeyBytes)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Could not parse public key from database!"))
    return
  }

  hasher := sha256.New()

  signData, err := ioutil.ReadAll(r.Body)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Could not read from body!"))
    return
  }

  hasher.Write(signData)

  signBytes, err := hex.DecodeString(signature)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Cannot decode hex data from signature parameter!"))
    return
  }

  err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hasher.Sum(nil), signBytes)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Signature did not verify!"))
    return
  }

  _, err = db.Exec("UPDATE cameras SET recording=? WHERE id=?;", status, cid)

  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Error updating camera record with status!"))
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Successfully updated camera record with status"))
}
