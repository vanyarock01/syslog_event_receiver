package syslog_event_receiver

import (
	"math"
	"time"
)

const severityMask = 7
const facilityShift = 3

var month = map[string]time.Month{
	"Jan": time.January,
	"Feb": time.February,
	"Mar": time.March,
	"Apr": time.April,
	"May": time.May,
	"Jun": time.June,
	"Jul": time.July,
	"Aug": time.August,
	"Sep": time.September,
	"Oct": time.October,
	"Nov": time.November,
	"Dec": time.December,
}

var monthIndexed = []time.Month{
	0,
	time.January,
	time.February,
	time.March,
	time.April,
	time.May,
	time.June,
	time.July,
	time.August,
	time.September,
	time.October,
	time.November,
	time.December,
}

// event is a parsed syslog event, validation of the format is done at the parser level.
type Event struct {
	message    string
	hostname   string //x
	priority   int
	program    string //x
	pid        int
	month      time.Month
	day        int
	hour       int
	minute     int
	second     int
	nanosecond int
	year       int
	loc        *time.Location
	sequence   int
}

// newEvent() return a new event.
func NewEvent() *Event {
	return &Event{
		priority: -1,
		pid:      -1,
		month:    -1,
		day:      -1,
		hour:     -1,
		minute:   -1,
		second:   -1,
		year:     time.Now().Year(),
		sequence: -1,
	}
}

// SetTimeZone set the timezone offset from the string.
func (s *Event) SetTimeZone(b []byte) {
	// We assume that we are in utc and ignore any other bytes after.
	// This can be followed by others bytes +00, +00:00 or +0000.
	if b[0] == 'Z' || b[0] == 'z' {
		s.loc = time.UTC
		return
	}

	d := 1
	if b[0] == '-' {
		d = -1
	}

	// +00 +00:00 or +0000
	// Use second value directly and don't use unecessary time.Duration.
	// Time.FixedZone accepts number of seconds.
	var h, m int
	switch len(b[1:]) {
	case 2:
		h = 3600 * bytesToInt(skipLeadZero(b[1:]))
		s.loc = time.FixedZone("", d*h)
	case 4:
		h = 3600 * bytesToInt(skipLeadZero(b[1:3]))
		m = 60 * bytesToInt(skipLeadZero(b[3:5]))
		s.loc = time.FixedZone("", d*(h+m))
	case 5:
		h = 3600 * bytesToInt(skipLeadZero(b[1:3]))
		m = 60 * bytesToInt(skipLeadZero(b[4:6]))
		s.loc = time.FixedZone("", d*(h+m))
	}
}

// SetMonthNumeric sets the month with a number.
func (s *Event) SetMonthNumeric(b []byte) {
	s.month = monthIndexed[bytesToInt(skipLeadZero(b))]
}

// SetMonth sets the month.
func (s *Event) SetMonth(b []byte) {
	var k string
	if len(b) > 3 {
		k = string(b[0:3])
	} else {
		k = string(b)
	}
	v, ok := month[k]
	if ok {
		s.month = v
	}
}

// Month returns the month.
func (s *Event) Month() time.Month {
	return s.month
}

// SetDay sets the day as.
func (s *Event) SetDay(b []byte) {
	s.day = bytesToInt(skipLeadZero(b))
}

// Day returns the day.
func (s *Event) Day() int {
	return s.day
}

// SetHour sets the hour.
func (s *Event) SetHour(b []byte) {
	s.hour = bytesToInt(skipLeadZero(b))
}

// Hour returns the hour.
func (s *Event) Hour() int {
	return s.hour
}

// SetMinute sets the minute.
func (s *Event) SetMinute(b []byte) {
	s.minute = bytesToInt(skipLeadZero(b))
}

// Minute return the minutes.
func (s *Event) Minute() int {
	return s.minute
}

// SetSecond sets the second.
func (s *Event) SetSecond(b []byte) {
	s.second = bytesToInt(skipLeadZero(b))
}

// Second returns the second.
func (s *Event) Second() int {
	return s.second
}

// SetYear sets the current year.
func (s *Event) SetYear(b []byte) {
	s.year = bytesToInt(b)
}

// Year returns the current year, since syslog events don't include that.
func (s *Event) Year() int {
	return s.year
}

// SetMessage sets the message.
func (s *Event) SetMessage(b []byte) {
	s.message = string(b)
}

// Message returns the message.
func (s *Event) Message() string {
	return s.message
}

// SetPriority sets the priority.
func (s *Event) SetPriority(priority []byte) {
	s.priority = bytesToInt(priority)
}

// Priority returns the priority.
func (s *Event) Priority() int {
	return s.priority
}

// HasPriority returns if the priority was in original event.
func (s *Event) HasPriority() bool {
	return s.priority >= 0
}

// Severity returns the severity, will return -1 if priority is not set.
func (s *Event) Severity() int {
	if !s.HasPriority() {
		return -1
	}
	return s.Priority() & severityMask
}

// Facility returns the facility, will return -1 if priority is not set.
func (s *Event) Facility() int {
	if !s.HasPriority() {
		return -1
	}
	return s.Priority() >> facilityShift
}

// SetHostname sets the hostname.
func (s *Event) SetHostname(b []byte) {
	s.hostname = string(b)
}

// Hostname returns the hostname.
func (s *Event) Hostname() string {
	return string(s.hostname)
}

// SetProgram sets the programs as a byte slice.
func (s *Event) SetProgram(b []byte) {
	s.program = string(b)
}

// Program returns the program name.
func (s *Event) Program() string {
	return s.program
}

func (s *Event) SetPid(b []byte) {
	s.pid = bytesToInt(b)
}

// Pid returns the pid.
func (s *Event) Pid() int {
	return s.pid
}

// HasPid returns true if a pid is set.
func (s *Event) HasPid() bool {
	return s.pid > 0
}

// SetSequence set the sequence number for this event.
func (s *Event) SetSequence(b []byte) {
	s.sequence = bytesToInt(b)
}

// Sequence returns the sequence number of the event when defined,
// otherwise return -1.
func (s *Event) Sequence() int {
	return s.sequence
}

// SetNanoSecond sets the nanosecond.
func (s *Event) SetNanosecond(b []byte) {
	// We assume that we receive a byte array representing a nanosecond, this might not be
	// always the case, so we have to pad it.
	if len(b) < 9 {
		s.nanosecond = bytesToInt(skipLeadZero(b)) * int(math.Pow10((9 - len(b))))
	} else {
		s.nanosecond = bytesToInt(skipLeadZero(b))
	}
}

// NanoSecond returns the nanosecond.
func (s *Event) Nanosecond() int {
	return s.nanosecond
}

// Timestamp return the timestamp in UTC.
func (s *Event) Timestamp(timezone *time.Location) time.Time {
	var t *time.Location
	if s.loc == nil {
		t = timezone
	} else {
		t = s.loc
	}

	return time.Date(
		s.Year(),
		s.Month(),
		s.Day(),
		s.Hour(),
		s.Minute(),
		s.Second(),
		s.Nanosecond(),
		t,
	).UTC()
}

// IsValid returns true if the date and the message are present.
func (s *Event) IsValid() bool {
	return s.day != -1 && s.hour != -1 && s.minute != -1 && s.second != -1 && s.message != ""
}

// BytesToInt takes a variable length of bytes and assume ascii chars and convert it to int, this is
// a simplified implementation of strconv.Atoi's fast path without error handling and remove the
// need to convert the byte array to string, we also assume that any errors are taken care at
// the parsing level.
func bytesToInt(b []byte) int {
	var i int
	for _, x := range b {
		i = i*10 + int(x-'0')
	}
	return i
}

func skipLeadZero(b []byte) []byte {
	if len(b) > 1 && b[0] == '0' {
		return b[1:]
	}
	return b
}
