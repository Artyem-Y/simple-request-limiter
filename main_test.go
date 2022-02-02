package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServeHTTP_GetNotSupported_ReturnError(t *testing.T) {
	t.Parallel()

	// arrange
	expected := "Sorry, method GET not supported"
	req, err := http.NewRequest("GET", "/urls", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := HttpHandler{}
	// act
	handler.ServeHTTP(res, req)

	// assert
	assert.Contains(t, res.Body.String(), expected)
}

func TestServeHTTP_UrlNotFound_ReturnError(t *testing.T) {
	t.Parallel()

	// arrange
	expected := "404 page not found"
	req, err := http.NewRequest("GET", "/urls-2", nil)
	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := HttpHandler{}
	// act
	handler.ServeHTTP(res, req)

	// assert
	assert.Contains(t, res.Body.String(), expected)
}

func TestServeHTTP_PostUrls_ReturnResultStatusOk(t *testing.T) {
	t.Parallel()

	// arrange
	type Expected struct {
		Data    string `json:"data"`
		Message string `json:"message"`
	}

	var jsonStr = []byte(`{
		"urls": "https://google.com\r\nhttps://google.com\r\nhttps://google.com\r\nhttps://google.com\r\nhttps://google.com"
	}`)
	req, err := http.NewRequest("POST", "/urls", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := HttpHandler{}
	// act
	handler.ServeHTTP(res, req)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var expected Expected
	err = json.Unmarshal(body, &expected)

	if err != nil {
		t.Fatal(err)
	}

	// assert
	assert.Equal(t, expected.Message, "Status OK")

}

func TestServeHTTP_PostEmptyBody_ReturnError(t *testing.T) {
	t.Parallel()

	// arrange
	expected := "Body can't be empty"
	var jsonStr = []byte(``)
	req, err := http.NewRequest("POST", "/urls", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := HttpHandler{}
	// act
	handler.ServeHTTP(res, req)

	// assert
	assert.Contains(t, res.Body.String(), expected)
}

func TestServeHTTP_PostErrorOccured_ReturnError(t *testing.T) {
	t.Parallel()

	// arrange
	expected := "Body can't be empty"
	req, err := http.NewRequest("POST", "/urls", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}
	res := httptest.NewRecorder()
	handler := HttpHandler{}
	// act
	handler.ServeHTTP(res, req)

	// assert
	assert.Contains(t, res.Body.String(), expected)
}

func TestSHA1(t *testing.T) {
	// arrange
	bodyString := "https://google.com"
	h := sha1.New()
	h.Write([]byte(bodyString))

	// act
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	// assert
	assert.Equal(t, sha1_hash, "72fe95c5576ec634e214814a32ab785568eda76a")
}
