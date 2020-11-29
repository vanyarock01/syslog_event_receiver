package main

import (
	// "errors"
	// "bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	// "strings"
	syslog "github.com/vanyarock01/syslog_event_receiver"
)

func httpServer(srv *syslog.UDPServer) {
	http.HandleFunc("/syslog-event-receiver/start", func(w http.ResponseWriter, rec *http.Request) {
		if rec.Method == "POST" {
			err := srv.Start()
			if err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				fmt.Fprint(w, "OK")
			}
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/syslog-event-receiver/stop", func(w http.ResponseWriter, rec *http.Request) {
		if rec.Method == "POST" {
			err := srv.Stop()
			if err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				fmt.Fprint(w, "OK")
			}
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})
	http.ListenAndServe(":8090", nil)
}

func main() {
	port := os.Getenv("SYSLOG_PORT")
	if port == "" {
		log.Print("[error] ENV variable SYSLOG_PORT not set")
	}

	events := make(chan []byte, 1024)
	srv := syslog.NewUdpServer("127.0.0.1", port, events)

	err := srv.Start()
	if err != nil {
		log.Printf("[error] %s", err)
		os.Exit(1)
	}

	go httpServer(srv)

	go func() {
		// wait few seconds
		time.Sleep(10 * time.Second)
		// stop server
		srv.Stop()
		log.Printf("[info] Server pause")
		time.Sleep(10 * time.Second)
		srv.Start()
		log.Printf("[info] Server start")
	}()

	for event := range events {
		log.Printf("[info] Receive event: %s", string(event))
	}
}
