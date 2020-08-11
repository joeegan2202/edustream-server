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
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// IngestServer : A server for handling file ingest from deploy server
type IngestServer struct{}

func (i *IngestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query()

	if query["signature"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect parameters! Missing signature"))
		return
	}

	signature := query["signature"][0]

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
		room            string
		sid             string
		publicKeyString string
		publicKey       *rsa.PublicKey
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

	bytesRead, err := io.ReadAtLeast(r.Body, signData, 100)

	if err != nil {
		logger.Printf("Error trying to read bytes for signing! %s\n", err.Error())
	}

	hasher := sha256.New()

	hasher.Write(signData[:bytesRead])

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

	os.MkdirAll(dir, 0755)
	if err != nil {
		logger.Printf("Error trying to remove file ingesting! %s\n", err.Error())
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, filename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Exec("UPDATE cameras SET lastStreamed=? WHERE sid=? AND id=?;", time.Now().Unix(), sid, cid)

	if err != nil {
		logger.Printf("Error updating camera record with streaming time! %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not update camera record with streaming time!")
	}

	var classID string
	err = db.QueryRow("SELECT classes.id FROM classes INNER JOIN cameras ON cameras.room=classes.room INNER JOIN periods ON periods.code=classes.period INNER JOIN schools ON schools.id=cameras.sid WHERE periods.stime<unix_timestamp() AND periods.etime>unix_timestamp() AND schools.id=? AND cameras.id=?;", sid, cid).Scan(&classID)

	if err != nil {
		logger.Printf("Error trying to get class id! %s\n", err.Error())
	}

	rows, err = db.Query("SELECT * FROM recording WHERE cid=? AND status=0;", classID)

	if err == nil {
		if !rows.Next() {
			db.Exec("INSERT INTO recording ( sid, cid, time, status ) VALUES ( ?, ?, unix_timestamp(), 0 );", sid, classID)
		}
	} else {
		logger.Printf("Error querying for if class is recording! %s\n", err.Error())
	}

	io.Copy(file, io.MultiReader(bytes.NewReader(signData[:bytesRead]), r.Body))
}
