package common

import "log"

// must panics in the case of error.
func Must(err error) {
	if err == nil {
		return
	}

	log.Panicln(err)
}
