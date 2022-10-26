package main

import (
	"fmt"
	"os"
	"os/signal"
)

func CatchOSSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for sig := range c {
		if sig == os.Interrupt {
			fmt.Println("")
			break
		}
	}
}
