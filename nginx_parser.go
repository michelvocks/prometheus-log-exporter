package main

import (
	"net/http"
	"strings"
	"fmt"
	"strconv"
	"time"
	"regexp"
)

type nginxData struct {
	method string
	bucket string
	status int
	count int
	lastreset time.Time
	time float64
}

const (
	IMAGE_BUCKET = "images"
	JS_BUCKET = "javascript"
	FONTS_BUCKET = "fonts"
	CSS_BUCKET = "css"
	DOCS_BUCKET = "documents"
	OTHER_BUCKET = "otherfiles"
	HTML_BUCKET = "html"
)

var (
	images = regexp.MustCompile("(.jpg|.jpeg|.png|.gif|.svg)$")
	js = regexp.MustCompile("(.js)$")
	fonts = regexp.MustCompile("(.woff|.woff2)$")
	css = regexp.MustCompile("(.css)$")
	docs = regexp.MustCompile("(.pdf|.pptx|.docx|.doc|.xls|.xlsx)$")
)

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
	path := s[8]

	// Find out the belonging bucket
	re := regexp.MustCompile("\\.(\\w{2,5})$")
	if re.FindStringIndex(path) != nil {
		// lets do it more precisely
		if images.FindStringIndex(path) != nil {
			nd.bucket = IMAGE_BUCKET
		} else if js.FindStringIndex(path) != nil {
			nd.bucket = JS_BUCKET
		} else if fonts.FindStringIndex(path) != nil {
			nd.bucket = FONTS_BUCKET
		} else if css.FindStringIndex(path) != nil {
			nd.bucket = CSS_BUCKET
		} else if docs.FindStringIndex(path) != nil {
			nd.bucket = DOCS_BUCKET
		} else {
			nd.bucket = OTHER_BUCKET
		}
	} else {
		nd.bucket = HTML_BUCKET
	}

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
		if ndc.bucket == nd.bucket && ndc.status == nd.status && ndc.method == nd.method {
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
	fmt.Fprint(w, "#TYPE nginx_request_time_seconds_avg gauge\n")

	// Iterate all entries
	for _, nd := range n {
		// Calculate the average
		avg := nd.time / float64(nd.count)

		// print out the average
		fmt.Fprintf(w, "nginx_request_time_seconds_avg{bucket=\"%s\",method=\"%s\",status=\"%d\"} %0.3f\n", nd.bucket, nd.method, nd.status, avg)
	}

}
