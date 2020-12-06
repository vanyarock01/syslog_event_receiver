package syslog_event_receiver

import (
	"fmt"
	"log"
	"time"

	tnt "github.com/tarantool/go-tarantool"
)

type DB interface {
	Connect(string, string, Opts) error
	Close() error
	InsertEvent(*Event) error
}

type Tuple []interface{}
type Opts map[string]interface{}

type TntDB struct {
	conn *tnt.Connection
}

func NewTntDB() *TntDB {
	return &TntDB{nil}
}

func (db *TntDB) Connect(host string, port string, opts Opts) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	conn, err := tnt.Connect(addr, tnt.Opts{
		User: opts["user"].(string),
		Pass: opts["pass"].(string),
	})

	if err != nil {
		return fmt.Errorf("can't connect to DB on '%s' <%s>", addr, err)
	}

	log.Printf("[info] connected to Tarantool DB on '%s'", addr)
	db.conn = conn

	return nil
}

func (db *TntDB) Close() error {
	return db.conn.Close()
}

func (db *TntDB) InsertEvent(event *Event) error {
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

	_, err := db.conn.Insert("syslog", t)

	return err
}
