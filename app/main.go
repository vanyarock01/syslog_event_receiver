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
		log.Print("[panic] ENV variable SYSLOG_PORT not set")
		os.Exit(1)
	}

	srv := syslog.NewSyslogServer("127.0.0.1", udpPort, "8081")

	err := srv.Start()
	if err != nil {
		log.Printf("[panic] %s", err)
		return
	}
	defer srv.Stop()

	srv.InitAPI()

	// Run storage
	err = srv.InitDBConn("127.0.0.1", "3301", syslog.Opts{"user": "gouser", "pass": "secret"})

	if err != nil {
		log.Printf("[panic] %s", err)
		return
	}
	defer srv.CloseDBConn()

	go func() {
		// wait few seconds
		time.Sleep(10 * time.Second)
		// stop server
		srv.Stop()
		log.Printf("[info] server pause")
		time.Sleep(10 * time.Second)
		srv.Start()
		log.Printf("[info] server start")
	}()

	for {
		/* do nothing */
	}

}
