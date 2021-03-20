package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

var (
	router    *mux.Router
	twitsChan chan stocktwits.Message
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws *utils.WebsocketMap
)

func dataPipe(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	ws.Set(fmt.Sprint(time.Now().UnixNano()), conn)
}

func buildRouter() {
	router.HandleFunc("/", dataPipe)
}

func main() {

	var stocktwitsChan = make(chan []stocktwits.Message, 64)

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configFilePath := flag.String("config-file-path", path.Join(home, "settings.json"), "a string")

	flag.Parse()

	utils.ReadConfig(*configFilePath)

	dataSources := utils.DataSources()

	fmt.Println(dataSources)

	go http.ListenAndServe(":8492", router)

	go stocktwits.SuggestedStream(stocktwitsChan, time.Minute)

	go func() {
		for {
			ws.Broadcast(<-stocktwitsChan)
		}
	}()

	utils.CloseHandler()

	return
}
