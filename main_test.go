package main

import (
	"testing"
	"os"
	"net/http"
	"net/http/httptest"
	"fmt"
)

const testLogString  = "127.0.0.1  http - - [24/May/2017:12:33:33 +0000] \"GET /global/about/press/science-based-targets-initiative-7498 HTTP/1.1\" 200 20797 0.046 \"https://wwwtest.dbschenker.com/global/6054!search?formState=eNoVirEKhDAQBX_leLWFttbXHFipXL8kexoIiZfdICL-u2v33syc-JFjFfTn1WCjJSTSkJN9lLwb79oGolT0mdb8K5cDfaoxmsgPhw-yRTrYv0n55VkczDEVt44sNeo38G7d8JlmXDcP1yeb&page=0\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393\""

func TestScrape(t *testing.T) {
	os.Setenv("LOG_PATH", "test/access-ssl.log,test/access-ssl2.log")
	setup()

	// Add new log line to file
	file, err := os.OpenFile("test/access-ssl.log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString(testLogString); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Tear up
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(metricsHandler)

	// call it
	handler.ServeHTTP(rr, req)

	// Check response
	fmt.Println(rr.Body.String())
}
