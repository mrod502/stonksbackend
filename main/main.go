package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	stocktwits "github.com/mrod502/stocktwitsgo"
	"github.com/mrod502/stonksbackend/utils"
)

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

	go http.ListenAndServe(fmt.Sprintf(":%d", utils.ServePort()), router)

	go stocktwits.SuggestedStream(stocktwitsChan, time.Minute)

	go func() {
		for {
			ws.Broadcast(<-stocktwitsChan)
		}
	}()

	utils.CloseHandler()

	return
}
