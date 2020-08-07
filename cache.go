package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// CachedFile type for the cached files
type CachedFile struct {
	path string
	data *[]byte
	time int64
}

var cache []*CachedFile
var cacheSize int

// MaxCache global variable for maximum cache size
var MaxCache int

func manageCache() {
	var err error
	MaxCache, err = strconv.Atoi(os.Getenv("MAX_CACHE"))

	if err != nil {
		logger.Fatalf("Error trying to get MAX_CACHE env variable! %s\n", err.Error())
	}

	cacheSize = 0

	cache = make([]*CachedFile, 0)

	for {
		timer := time.After(1 * time.Second)

		for i, file := range cache {
			if file.time > time.Now().Unix()-60 {
				continue
			}

			cache = append(cache[:i], cache[i+1:]...)
		}

		accumulator := 0

		for _, file := range cache {
			accumulator += len(*file.data)
		}

		cacheSize = accumulator

		<-timer
	}
}

func insertCache(path string, file io.Reader) error {
	if cacheSize >= MaxCache {
		io.Copy(ioutil.Discard, file)
		return nil
	}

	data, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	if !(len(data) > 0) {
		return fmt.Errorf("empty file. Not written to cache")
	}

	cfile := CachedFile{path, &data, time.Now().Unix()}
	cache = append(cache, &cfile)
	return nil
}
