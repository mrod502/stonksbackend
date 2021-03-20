package utils

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//CloseHandler - wait for ctrl c to close
func CloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("ctrl+c pressed, exiting")
}
