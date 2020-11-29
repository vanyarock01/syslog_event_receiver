package syslog_event_receiver

import (
	"fmt"
	"log"
	"net"
)

type UDPServer struct {
	addr   string
	signal chan int
	conn   *net.UDPConn
	events chan<- []byte
}

const (
	STOP_SIGNAL = iota
)

func NewUdpServer(host string, port string, events chan<- []byte) *UDPServer {
	return &UDPServer{
		addr:   fmt.Sprintf("%s:%s", host, port),
		signal: make(chan int),
		events: events,
	}
}

func (srv *UDPServer) connect() error {
	udpAddr, err := net.ResolveUDPAddr("udp", srv.addr)
	if err != nil {
		return fmt.Errorf("can't resolve UDP addr '%s': %s", srv.addr, err)
	}

	srv.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("can't listen UDP: %s", err)
	}

	log.Printf("[info] Listen UDP on '%s'", srv.addr)
	return nil
}

func (srv *UDPServer) closeConnect() {
	srv.conn.Close()
	srv.conn = nil
}

func (srv *UDPServer) loop() {
	go func() {
		defer srv.closeConnect()
		log.Printf("[info] Begin event loop")

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
					log.Printf("[info] End event loop")
					return
				}
			default:
			}

			srv.events <- buffer
		}
	}()
}

func (srv *UDPServer) Start() error {
	if srv.conn != nil {
		return fmt.Errorf("Connection to '%s' already exist", srv.addr)
	}

	err := srv.connect()
	if err != nil {
		return err
	}

	srv.loop()
	return nil
}

func (srv *UDPServer) Stop() error {
	if srv.conn == nil {
		return fmt.Errorf("Connection to '%s' not exist", srv.addr)
	}
	srv.signal <- STOP_SIGNAL
	return nil
}
