package syslog_event_receiver

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
)

type SyslogServer struct {
	udpAddr  string
	httpAddr string
	signal   chan int
	conn     *net.UDPConn
	connMu   sync.Mutex
	db       DB
}

const (
	STOP_SIGNAL = iota
)

func NewSyslogServer(host string, udpPort string, httpPort string) *SyslogServer {
	return &SyslogServer{
		udpAddr:  fmt.Sprintf("%s:%s", host, udpPort),
		httpAddr: fmt.Sprintf("%s:%s", host, httpPort),
		signal:   make(chan int),
		db:       NewTntDB(),
	}
}

func (srv *SyslogServer) connect() error {
	udpAddr, err := net.ResolveUDPAddr("udp", srv.udpAddr)
	if err != nil {
		return fmt.Errorf("can't resolve UDP addr '%s' <%s>", srv.udpAddr, err)
	}

	srv.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("can't listen UDP <%s>", err)
	}

	log.Printf("[info] listen UDP on '%s'", srv.udpAddr)
	return nil
}

func (srv *SyslogServer) closeConnect() {
	srv.connMu.Lock()
	err := srv.conn.Close()
	srv.conn = nil
	srv.connMu.Unlock()

	if err != nil {
		log.Printf("[error] can't close udp connection <%s>", err)
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
				log.Printf("[error] can't read message from UDP <%s>", err)
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
			/* send events to storage
			may be use another gorutine for processing logs in background
			but tarantool pretty fast ^_^ */
			buffer = bytes.Trim(buffer, "\x00")
			event := NewEvent()
			Parse(buffer, event)

			log.Printf("[info] event message <%s>", event.Message())

			err = srv.db.InsertEvent(event)
			if err != nil {
				log.Printf("[error] can't insert log event to db <%s>", err)
			}
		}
	}()
}

func (srv *SyslogServer) Start() error {
	srv.connMu.Lock()
	if srv.conn != nil {
		return fmt.Errorf("connection to '%s' already exist", srv.udpAddr)
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
		err = fmt.Errorf("connection to '%s' not exist", srv.udpAddr)
	}
	srv.connMu.Unlock()

	if err != nil {
		return err
	}

	srv.signal <- STOP_SIGNAL
	return nil
}

func (srv *SyslogServer) InitDBConn(host string, port string, opts map[string]interface{}) error {
	err := srv.db.Connect(host, port, opts)
	if err != nil {
		return err
	}

	return nil
}

func (srv *SyslogServer) CloseDBConn() {
	err := srv.db.Close()
	if err != nil {
		log.Printf("[error] can't close connection to DB <%s>", err)
	}
}
