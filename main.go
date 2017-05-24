package main

import (
	"net/http"
	"os"
	"log"
	"bufio"
	"fmt"
	"sync"
	"strings"
)

type fileHandler struct {
	pos int64
	path string
	mutex sync.RWMutex
}

// storePos stores safely the position
func (file *fileHandler) storePos(givenPos int64) {
	// Lock this object before we change it
	// Also unlock it after we leave the function
	file.mutex.Lock()
	defer file.mutex.Unlock()

	file.pos = givenPos
}

var fileHandlers[] fileHandler

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	for id := range fileHandlers {
		// Check the offset position
		fileHandlers[id].estimateStart()

		fileDef, err := os.Open(fileHandlers[id].path)
		if err != nil {
			log.Fatal(err)
		}
		defer fileDef.Close()

		if _, err = fileDef.Seek(fileHandlers[id].pos, 0); err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(fileDef)

		pos := fileHandlers[id].pos
		scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = bufio.ScanLines(data, atEOF)
			pos += int64(advance)
			return
		}
		scanner.Split(scanLines)

		for scanner.Scan() {
			// Analyze line by line

			fmt.Fprintf(w, "Pos: %d, Scanned: %s\n", pos, scanner.Text())
		}

		// print type
		fmt.Fprintf(w, "# TYPE nginx_request_time_microseconds_avg gauge")

		// Set the new position and do it threadsafe
		fileHandlers[id].storePos(pos)
	}
}

func (file *fileHandler) estimateStart() {
	stat, err := os.Stat(file.path)
	if err != nil {
		log.Panic(err)
	}

	// Check if chars has been deleted
	if stat.Size() < file.pos {
		// Always set it to the last line
		file.pos = stat.Size()
	}

	// Set start to end of file
	if file.pos == 0 {
		file.pos = stat.Size()
	}
}

func setup() {
	// Get file list
	fileList := strings.Split(os.Getenv("LOG_PATH"), ",")

	// Iterate file list and add new objects with path
	for _, filePath := range fileList {
		singleFile := fileHandler{}
		singleFile.path = filePath
		singleFile.estimateStart()
		fileHandlers = append(fileHandlers, singleFile)
	}
}

func main() {
	setup()

	http.HandleFunc("/metrics", metricsHandler)
	http.ListenAndServe(":8080", nil)
}


