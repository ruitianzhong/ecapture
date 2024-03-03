package event_processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	testFile = "testdata/all.json"
)

type SSLDataEventTmp struct {
	//Event_type   uint8    `json:"Event_type"`
	DataType  int64      `json:"DataType"`
	Timestamp uint64     `json:"Timestamp"`
	Pid       uint32     `json:"Pid"`
	Tid       uint32     `json:"Tid"`
	Data_len  int32      `json:"DataLen"`
	Comm      [16]byte   `json:"Comm"`
	Fd        uint32     `json:"Fd"`
	Version   int32      `json:"Version"`
	Data      [4096]byte `json:"Data"`
}

func TestEventProcessor_Serve(t *testing.T) {

	logger := log.Default()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	/*
		f, e := os.Create("./output.log")
		if e != nil {
			t.Fatal(e)
		}
		logger.SetOutput(f)
	*/
	ep := NewEventProcessor(logger, true)

	go func() {
		ep.Serve()
	}()
	content, err := os.ReadFile(testFile)
	if err != nil {
		//Do something
		log.Fatalf("open file error: %s, file:%s", err.Error(), testFile)
	}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var eventSSL SSLDataEventTmp
		err := json.Unmarshal([]byte(line), &eventSSL)
		if err != nil {
			t.Fatalf("json unmarshal error: %s, body:%v", err.Error(), line)
		}
		payloadFile := fmt.Sprintf("testdata/%d.bin", eventSSL.Timestamp)
		b, e := os.ReadFile(payloadFile)
		if e != nil {
			t.Fatalf("read payload file error: %s, file:%s", e.Error(), payloadFile)
		}
		copy(eventSSL.Data[:], b)
		ep.Write(&BaseEvent{Data_len: eventSSL.Data_len, Data: eventSSL.Data, DataType: eventSSL.DataType, Timestamp: eventSSL.Timestamp, Pid: eventSSL.Pid, Tid: eventSSL.Tid, Comm: eventSSL.Comm, Fd: eventSSL.Fd, Version: eventSSL.Version})
	}

	tick := time.NewTicker(time.Second * 3)
	<-tick.C

	err = ep.Close()
	lines = strings.Split(buf.String(), "\n")
	ok := true
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "dump") {
			t.Log(line)
			ok = false
		}
	}
	if err != nil {
		t.Fatalf("close error: %s", err.Error())
	}
	if !ok {
		t.Fatalf("some errors occurred")
	}
	t.Log(buf.String())
	t.Log("done")
}
// go test  -v ./pkg/event_processor/  -run TestHang 
func TestHang(t *testing.T) {

	logger := log.Default()
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	ep := NewEventProcessor(logger, true)

	go func() {
		ep.Serve()
	}()
	content, err := os.ReadFile(testFile)
	if err != nil {
		//Do something
		log.Fatalf("open file error: %s, file:%s", err.Error(), testFile)
	}
	lines := strings.Split(string(content), "\n")
	line := lines[0]
	var eventSSL SSLDataEventTmp
	json.Unmarshal([]byte(line), &eventSSL)
	payloadFile := fmt.Sprintf("testdata/%d.bin", eventSSL.Timestamp)
	b, e := os.ReadFile(payloadFile)
	fmt.Println(string(b))
	if e != nil {
		t.Fatalf("read payload file error: %s, file:%s", e.Error(), payloadFile)
	}
	copy(eventSSL.Data[:], b)
	ep.Write(&BaseEvent{Data_len: eventSSL.Data_len, Data: eventSSL.Data, DataType: eventSSL.DataType, Timestamp: eventSSL.Timestamp, Pid: eventSSL.Pid, Tid: eventSSL.Tid, Comm: eventSSL.Comm, Fd: eventSSL.Fd, Version: eventSSL.Version})
	time.Sleep(2 * time.Second)
	// stress test mocking the same process
	for i := 0; i < 3000; i++ {
		ep.Write(&BaseEvent{Data_len: eventSSL.Data_len, Data: eventSSL.Data, DataType: eventSSL.DataType, Timestamp: eventSSL.Timestamp, Pid: eventSSL.Pid, Tid: eventSSL.Tid, Comm: eventSSL.Comm, Fd: eventSSL.Fd, Version: eventSSL.Version})
	}
	fmt.Println("no hang")
	time.Sleep(time.Second * 20)
	fmt.Println("end")

	ep.Close()
}
