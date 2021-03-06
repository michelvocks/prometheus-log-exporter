package main

import (
	"testing"
	"os"
	"net/http"
	"net/http/httptest"
	"fmt"
)

const testLogString  = "\n127.0.0.1  http - - [24/May/2017:12:33:33 +0000] \"GET /global/about/press/science-based-targets-initiative-7498 HTTP/1.1\" 200 20797 0.046 \"https://wwwtest.dbschenker.com/global/6054!search?formState=eNoVirEKhDAQBX_leLWFttbXHFipXL8kexoIiZfdICL-u2v33syc-JFjFfTn1WCjJSTSkJN9lLwb79oGolT0mdb8K5cDfaoxmsgPhw-yRTrYv0n55VkczDEVt44sNeo38G7d8JlmXDcP1yeb&page=0\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393\""
const testLogString2 = "\n127.0.0.1  http - - [24/May/2017:12:33:33 +0000] \"GET /global/about/press/science-based-targets-initiative-7498 HTTP/1.1\" 200 20797 0.086 \"https://wwwtest.dbschenker.com/global/6054!search?formState=eNoVirEKhDAQBX_leLWFttbXHFipXL8kexoIiZfdICL-u2v33syc-JFjFfTn1WCjJSTSkJN9lLwb79oGolT0mdb8K5cDfaoxmsgPhw-yRTrYv0n55VkczDEVt44sNeo38G7d8JlmXDcP1yeb&page=0\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393\""
const testLogString3 = "\n127.0.0.1  http - - [31/May/2017:09:54:19 +0000] \"GET /resource/image/196/stage-desktop/1024/256/55f49a30078e366f3a463662917f84f3/Rp/stagepicture1.jpg HTTP/1.1\" 304 0 0.002 \"https://www.dbschenker.com/global/products/ocean-freight/fcl-deliveries\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36\""
const resultTest = "#TYPE nginx_request_time_seconds_avg gauge\nnginx_request_time_seconds_avg{bucket=\"html\",method=\"GET\",status=\"200\"} 0.046\n"
const resultTest2 = "#TYPE nginx_request_time_seconds_avg gauge\nnginx_request_time_seconds_avg{bucket=\"html\",method=\"GET\",status=\"200\"} 0.059\n"
const resultTest3 = "#TYPE nginx_request_time_seconds_avg gauge\nnginx_request_time_seconds_avg{bucket=\"html\",method=\"GET\",status=\"200\"} 0.059\nnginx_request_time_seconds_avg{bucket=\"images\",method=\"GET\",status=\"304\"} 0.002\n"

func init() {
	setup("example_config.json")
}

func TestScrape(t *testing.T) {
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

	// Print response
	fmt.Println(rr.Body.String())

	// Check response
	if rr.Body.String() != resultTest {
		t.Errorf("wrong result: got %v expected %v", rr.Body.String(), resultTest)
	}
}

func TestScrapeMulti(t *testing.T) {
	// Add new log line to file
	file, err := os.OpenFile("test/access-ssl.log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString(testLogString); err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(testLogString2); err != nil {
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

	// Print response
	fmt.Println(rr.Body.String())

	// Check response
	if rr.Body.String() != resultTest2 {
		t.Errorf("wrong result: got %v expected %v", rr.Body.String(), resultTest2)
	}
}

func TestScrapeImages(t *testing.T) {
	// Add new log line to file
	file, err := os.OpenFile("test/access-ssl.log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString(testLogString3); err != nil {
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

	// Print response
	fmt.Println(rr.Body.String())

	// Check response
	if rr.Body.String() != resultTest3 {
		t.Errorf("wrong result: got %v expected %v", rr.Body.String(), resultTest3)
	}
}
