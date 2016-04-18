package adapter

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	a := New(
		"notreal",
		"development,sandbox",
		1,
	)

	if a.logglyURL == "" {
		t.Error("expected New to set logglyURL")
	}
}

func TestBuffer(t *testing.T) {
	bufferSize := 10
	a := &Adapter{
		bufferSize: bufferSize,
	}
	buf := a.newBuffer()

	if len(buf) != 0 {
		t.Errorf("expected new buffer length to be 0 but it was %s", len(buf))
	}

	if cap(buf) != bufferSize {
		t.Errorf("expected buffer capacity to be %s but it was %s", bufferSize, cap(buf))
	}

	buf = append(buf, logglyMessage{})

	if len(buf) != 1 {
		t.Errorf(
			"expected buffer length to be 1 after adding item but it was %s",
			len(buf),
		)
	}

	oldBuffer := buf
	newBuffer := a.newBuffer()

	if &oldBuffer == &newBuffer {
		t.Error("expected oldBuffer and newBuffer to point to different structs")
	}
}

func TestBuildLogglyURL(t *testing.T) {
	var expectedURL string
	var actualURL string

	tag := "development"
	tags := "development,sandbox"
	token := "notreal"

	expectedURL = "https://logs-01.loggly.com/bulk/notreal"
	actualURL = buildLogglyURL(token, "")

	if actualURL != expectedURL {
		t.Errorf(
			"expected URL to be %s but was %s",
			expectedURL,
			actualURL,
		)
	}

	expectedURL = "https://logs-01.loggly.com/bulk/notreal/tag/development/"
	actualURL = buildLogglyURL(token, tag)

	if actualURL != expectedURL {
		t.Errorf(
			"expected actualURL to be %s but was %s",
			expectedURL,
			actualURL,
		)
	}

	expectedURL = "https://logs-01.loggly.com/bulk/notreal/tag/development,sandbox/"
	actualURL = buildLogglyURL(token, tags)

	if actualURL != expectedURL {
		t.Errorf(
			"expected actualURL to be %s but was %s",
			expectedURL,
			actualURL,
		)
	}
}

func TestSendRequestToLoggly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	var testBuff bytes.Buffer

	a := &Adapter{
		bufferSize: 1,
		log:        log.New(&testBuff, "logspout-loggly", log.LstdFlags),
		logglyURL:  ts.URL,
		queue:      make(chan logglyMessage),
	}

	req, err := http.NewRequest(
		"POST",
		ts.URL,
		bytes.NewBufferString("fake data"),
	)

	if err != nil {
		t.Errorf("expected error to be nil when creating request: %s", err.Error())
	}

	a.sendRequestToLoggly(req)

	if !(testBuff.Len() > 0) {
		t.Error("expected snedRequestToLoggly to write to log when it receives a non 200 response from loggly")
	}
}
