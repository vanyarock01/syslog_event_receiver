package syslog_event_receiver

import (
	"fmt"
	"net/http"
)

// HTTP API
func (srv *SyslogServer) startHandler(w http.ResponseWriter, rec *http.Request) {
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
}

func (srv *SyslogServer) stopHandler(w http.ResponseWriter, rec *http.Request) {
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
}

func (srv *SyslogServer) InitAPI() {
	go func() {
		http.HandleFunc("/syslog/start", srv.startHandler)
		http.HandleFunc("/syslog/stop", srv.stopHandler)
		http.ListenAndServe(srv.httpAddr, nil)
	}()
}
