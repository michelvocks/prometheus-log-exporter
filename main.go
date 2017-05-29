package main

import (
	"net/http"
	"os"
	"log"
	"bufio"
	"sync"
	"encoding/json"
)

type config struct {
	Nginx []string
}

const (
	nginx int = iota
)

var Configuration config

type fileHandler struct {
	pos int64
	path string
	logtype int
	mutex sync.RWMutex
}

type parser interface {
	parse(l string)
	print(w http.ResponseWriter)
}

var nginxCol parser = &nginx_col{}

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

		// Iterate over all new log file lines and parse them
		for scanner.Scan() {
			nginxCol.parse(scanner.Text())
		}

		// print the result
		nginxCol.print(w)

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

func setup(configPath string) {
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Decode json config file
	decoder := json.NewDecoder(file)
	Configuration = config{}
	err = decoder.Decode(&Configuration)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate nginx file path list and add new objects with path
	for _, nginxFile := range Configuration.Nginx {
		singleFile := fileHandler{}
		singleFile.path = nginxFile
		singleFile.logtype = nginx
		singleFile.estimateStart()
		fileHandlers = append(fileHandlers, singleFile)
	}
}

func main() {
	setup(os.Args[1])

	http.HandleFunc("/metrics", metricsHandler)
	http.ListenAndServe(":9011", nil)
}


