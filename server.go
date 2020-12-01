package syslog_event_receiver

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

// like documentation
type SyslogServerInterface interface {
	Start() error
	Stop() error
	InitAPI()

	connect() error
	closeConnect()
	loop()
}

type SyslogServer struct {
	udpAddr  string
	httpAddr string
	signal   chan int
	conn     *net.UDPConn
	connMu   sync.Mutex
	events   chan<- []byte
}

const (
	STOP_SIGNAL = iota
)

func NewSyslogServer(host string, udpPort string, httpPort string, events chan<- []byte) *SyslogServer {
	return &SyslogServer{
		udpAddr:  fmt.Sprintf("%s:%s", host, udpPort),
		httpAddr: fmt.Sprintf("%s:%s", host, httpPort),
		signal:   make(chan int),
		events:   events,
	}
}

func (srv *SyslogServer) connect() error {
	udpAddr, err := net.ResolveUDPAddr("udp", srv.udpAddr)
	if err != nil {
		return fmt.Errorf("can't resolve UDP addr '%s': %s", srv.udpAddr, err)
	}

	srv.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("can't listen UDP: %s", err)
	}

	log.Printf("[info] Listen UDP on '%s'", srv.udpAddr)
	return nil
}

func (srv *SyslogServer) closeConnect() {
	srv.connMu.Lock()
	err := srv.conn.Close()
	srv.conn = nil
	srv.connMu.Unlock()

	if err != nil {
		log.Printf("[error] can't close connection: %s", err)
	}
}

func (srv *SyslogServer) loop() {
	go func() {
		defer srv.closeConnect()
		log.Printf("[info] begin event loop")

		buffer := make([]byte, 1024)
		for {
			_, _, err := srv.conn.ReadFromUDP(buffer)
			if err != nil {
				log.Printf("[error] can't read message from UDP: %s", err)
				continue
			}
			// check STOP signal
			select {
			case signal := <-srv.signal:
				if signal == STOP_SIGNAL {
					log.Printf("[info] end event loop")
					return
				}
			default:
			}
			// send events to client
			srv.events <- buffer
		}
	}()
}

func (srv *SyslogServer) Start() error {
	srv.connMu.Lock()
	if srv.conn != nil {
		return fmt.Errorf("Connection to '%s' already exist", srv.udpAddr)
	}

	err := srv.connect()
	srv.connMu.Unlock()

	if err != nil {
		return err
	}

	srv.loop()
	return nil
}

func (srv *SyslogServer) Stop() (err error) {
	srv.connMu.Lock()
	if srv.conn == nil {
		err = fmt.Errorf("Connection to '%s' not exist", srv.udpAddr)
	}
	srv.connMu.Unlock()

	if err != nil {
		return err
	}

	srv.signal <- STOP_SIGNAL
	return nil
}

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
