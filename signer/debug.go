package signer

import "log"

var debugEnabled = false

// SetDebug enabled or disabled
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

func debugf(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf(format, v...)
	}
}
func debugln(v ...interface{}) {
	if debugEnabled {
		log.Println(v...)
	}
}
