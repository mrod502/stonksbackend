package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	stocktwits "github.com/mrod502/stocktwitsgo"
	"github.com/mrod502/stonksbackend/utils"
)

var (
	router    *mux.Router
	twitsChan chan stocktwits.Message
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws *utils.WebsocketMap
)

func stocktwitsMessages(w http.ResponseWriter, r *http.Request) {}

func connUpgrader(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	ws.Set(fmt.Sprint(time.Now().UnixNano()), conn)
}

func buildRouter() {
	router.HandleFunc("/", stocktwitsMessages)
}
