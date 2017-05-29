package main

import (
	"net/http"
	"strings"
	"fmt"
	"strconv"
	"time"
)

type nginxData struct {
	method string
	path string
	status int
	count int
	lastreset time.Time
	time float64
}

type nginx_col []nginxData

func (n *nginx_col) parse(l string) {
	// create new object
	nd := nginxData{}

	// Split through spaces
	s := strings.Split(l, " ")

	// Extra check for blank lines
	if len(s) < 12 {
		return
	}

	// Get the method and trim quotes
	nd.method = strings.TrimPrefix(s[7], "\"")

	// Get path
	nd.path = s[8]

	// Get status
	i, err := strconv.Atoi(s[10])
	if err != nil {
		return
	}
	nd.status = i

	// Get response time
	f, err := strconv.ParseFloat(s[12], 64)
	if err != nil {
		return
	}
	nd.time = f

	// Set last reset time and count
	nd.lastreset = time.Now()
	nd.count++

	// Iterate over all pre-existing elements
	// and check if the obj already exists
	foundObj := false
	for id, ndc := range *n {
		if ndc.path == nd.path && ndc.status == nd.status && ndc.method == nd.method {
			(*n)[id].count++
			(*n)[id].time += nd.time
			foundObj = true

			// Every 60 seconds we want to reset the average
			duration := time.Since(ndc.lastreset)
			if duration.Seconds() >= 60 {
				(*n)[id].time = ndc.time / float64(ndc.count)
				(*n)[id].count = 1
				(*n)[id].lastreset = time.Now()
			}
		}
	}

	// Add object to list if we cannot find it
	if !foundObj {
		*n = append(*n, nd)
	}
}

func (n nginx_col) print(w http.ResponseWriter) {
	// print type for request average request time
	fmt.Fprint(w, "#TYPE nginx_request_time_microseconds_avg gauge\n")

	// Iterate all entries
	for _, nd := range n {
		// Calculate the average
		avg := nd.time / float64(nd.count)

		// print out the average
		fmt.Fprintf(w, "nginx_request_time_microseconds_avg{path=\"%s\",method=\"%s\",status=\"%d\"} %0.3f\n", nd.path, nd.method, nd.status, avg)
	}

}
