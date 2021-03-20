package main

import (
	"time"

	"github.com/gorilla/mux"
	stocktwits "github.com/mrod502/stocktwitsgo"
	"github.com/mrod502/stonksbackend/utils"
)

func init() {
	var err error
	router = new(mux.Router)
	twitsChan = make(chan stocktwits.Message, 128)
	ws, err = utils.NewWebsocketMap(15*time.Minute, true)
	if err != nil {
		panic(err)
	}
}
