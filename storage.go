package syslog_event_receiver

import (
	"fmt"
	"log"
	"time"

	tnt "github.com/tarantool/go-tarantool"
)

type (
	Tuple = []interface{}
)

func (srv *SyslogServer) InitDBConn(host string, port string, user string, pass string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := tnt.Connect(addr, tnt.Opts{User: user, Pass: pass})

	if err != nil {
		return fmt.Errorf("can't connect to DB on '%s' <%s>", addr, err)
	}

	log.Printf("[info] connected to Tarantool DB on '%s'", addr)
	srv.dbConn = conn

	return nil
}

func (srv *SyslogServer) CloseDBConn() {
	err := srv.dbConn.Close()
	if err != nil {
		log.Printf("[error] can't close connection to DB <%s>", err)
	} else {
		log.Printf("[info] close connection to DB")
	}
}

func (srv *SyslogServer) insertEvent(event *Event) error {
	defaultLoc, _ := time.LoadLocation("UTC")

	t := Tuple{
		nil,                                    // auto id
		event.Timestamp(defaultLoc).UnixNano(), // required timestamp
		event.Message(),                        // may be empty, why not
		event.Hostname(),
		event.Priority(),
		event.Program(),
		event.Pid(),
		event.Sequence(),
	}
	log.Printf("[debug] insert tuple <%v>", t)

	_, err := srv.dbConn.Insert("syslog", t)

	return err
}
