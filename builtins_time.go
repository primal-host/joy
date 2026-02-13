package main

import (
	"fmt"
	"strings"
	"time"
)

func init() {
	// localtime: I -> L — Unix timestamp to local time list
	// List format: [year month day hour minute second isdst yearday weekday]
	register("localtime", func(m *Machine) {
		m.NeedStack(1, "localtime")
		a := m.Pop()
		t := time.Unix(a.Int, 0)
		m.Push(timeToList(t))
	})

	// gmtime: I -> L — Unix timestamp to UTC time list
	register("gmtime", func(m *Machine) {
		m.NeedStack(1, "gmtime")
		a := m.Pop()
		t := time.Unix(a.Int, 0).UTC()
		m.Push(timeToList(t))
	})

	// mktime: L -> I — time list to Unix timestamp
	register("mktime", func(m *Machine) {
		m.NeedStack(1, "mktime")
		a := m.Pop()
		if a.Typ != TypeList || len(a.List) < 6 {
			joyErr("mktime: time list with at least 6 elements expected")
		}
		year := int(a.List[0].Int)
		month := time.Month(a.List[1].Int)
		day := int(a.List[2].Int)
		hour := int(a.List[3].Int)
		min := int(a.List[4].Int)
		sec := int(a.List[5].Int)
		t := time.Date(year, month, day, hour, min, sec, 0, time.Local)
		m.Push(IntVal(t.Unix()))
	})

	// strftime: L S -> S2 — format time list with C-style format string
	register("strftime", func(m *Machine) {
		m.NeedStack(2, "strftime")
		fmtStr := m.Pop()
		tList := m.Pop()
		if fmtStr.Typ != TypeString {
			joyErr("strftime: format string expected")
		}
		if tList.Typ != TypeList || len(tList.List) < 6 {
			joyErr("strftime: time list expected")
		}
		year := int(tList.List[0].Int)
		month := time.Month(tList.List[1].Int)
		day := int(tList.List[2].Int)
		hour := int(tList.List[3].Int)
		min := int(tList.List[4].Int)
		sec := int(tList.List[5].Int)
		t := time.Date(year, month, day, hour, min, sec, 0, time.Local)
		m.Push(StringVal(strftime(fmtStr.Str, t)))
	})
}

func timeToList(t time.Time) Value {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7 (Mon=1..Sun=7)
	}
	_, offset := t.Zone()
	isDST := 0
	if offset != 0 {
		// Approximate DST detection: check if current offset differs from January offset
		jan := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
		_, janOffset := jan.Zone()
		if offset != janOffset {
			isDST = 1
		}
	}
	return ListVal([]Value{
		IntVal(int64(t.Year())),
		IntVal(int64(t.Month())),
		IntVal(int64(t.Day())),
		IntVal(int64(t.Hour())),
		IntVal(int64(t.Minute())),
		IntVal(int64(t.Second())),
		IntVal(int64(isDST)),
		IntVal(int64(t.YearDay())),
		IntVal(int64(weekday)),
	})
}

var weekdayNames = [...]string{
	"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
}

var weekdayShort = [...]string{
	"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat",
}

var monthNames = [...]string{
	"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

var monthShort = [...]string{
	"", "Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

func strftime(format string, t time.Time) string {
	var b strings.Builder
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' && i+1 < len(runes) {
			i++
			switch runes[i] {
			case 'Y':
				fmt.Fprintf(&b, "%04d", t.Year())
			case 'm':
				fmt.Fprintf(&b, "%02d", t.Month())
			case 'd':
				fmt.Fprintf(&b, "%02d", t.Day())
			case 'H':
				fmt.Fprintf(&b, "%02d", t.Hour())
			case 'M':
				fmt.Fprintf(&b, "%02d", t.Minute())
			case 'S':
				fmt.Fprintf(&b, "%02d", t.Second())
			case 'A':
				b.WriteString(weekdayNames[t.Weekday()])
			case 'a':
				b.WriteString(weekdayShort[t.Weekday()])
			case 'B':
				b.WriteString(monthNames[t.Month()])
			case 'b':
				b.WriteString(monthShort[t.Month()])
			case 'p':
				if t.Hour() < 12 {
					b.WriteString("AM")
				} else {
					b.WriteString("PM")
				}
			case 'Z':
				name, _ := t.Zone()
				b.WriteString(name)
			case '%':
				b.WriteByte('%')
			default:
				b.WriteByte('%')
				b.WriteRune(runes[i])
			}
		} else {
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}
