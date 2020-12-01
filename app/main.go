package main

import (
	// "fmt"
	"log"
	"os"
	"time"

	syslog "github.com/vanyarock01/syslog_event_receiver"
)

func main() {
	udpPort := os.Getenv("SYSLOG_PORT")
	if udpPort == "" {
		log.Print("[error] ENV variable SYSLOG_PORT not set")
		os.Exit(1)
	}

	events := make(chan []byte, 1024)
	srv := syslog.NewSyslogServer("127.0.0.1", udpPort, "8081", events)

	err := srv.Start()
	if err != nil {
		log.Printf("[error] %s", err)
		os.Exit(1)
	}

	srv.InitAPI()

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
