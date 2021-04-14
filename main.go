package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mrod502/logger"
	stocktwits "github.com/mrod502/stocktwitsgo"
	"github.com/mrod502/stonksbackend/utils"
)

func init() {
	var err error
	router = new(mux.Router)
	ws, err = utils.NewWebsocketMap(15*time.Minute, true)
	redditMessages = utils.NewRedditCache()
	twitsMessages = utils.NewStocktwitsCache()
	homeDir, _ = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
}
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	http.ServeFile(w, r, path.Join(homeDir, "favicon.ico"))
}

var (
	homeDir       string
	router        *mux.Router
	twitsMessages *utils.StocktwitsCache

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws             *utils.WebsocketMap
	redditMessages *utils.RedditCache
)

func dataPipe(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade", err.Error())
		return
	}
	ws.Set(fmt.Sprint(time.Now().UnixNano()), conn)
}
func stocktwitsMessages(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	b, err := json.Marshal(twitsMessages.All())
	if err != nil {
		logger.Error("twits", fmt.Sprintf("error unmarshaling twits messages: %s", err))
	}
	w.Write(b)
}

func tendies(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	b, err := json.Marshal(redditMessages.All())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
		return
	}
	w.Write(b)
}

func buildRouter() {
	router.HandleFunc("/ws", dataPipe)
	router.HandleFunc("/v1/stocktwits", stocktwitsMessages).Methods("GET")
	router.HandleFunc("/v1/reddit", tendies).Methods("GET")
	router.HandleFunc("/robots.txt", func(h http.ResponseWriter, r *http.Request) { h.Write([]byte("User-agent: *\nfuck off\n")) })
	router.HandleFunc("/favicon.ico", faviconHandler)
}

func main() {

	var stocktwitsChan = make(chan []stocktwits.Message, 64)

	home, err := os.UserHomeDir()

	if err != nil {
		panic(err)
	}

	configFilePath := flag.String("config-file-path", path.Join(home, "stonksbackend.json"), "a string")
	utils.ReadConfig(*configFilePath)

	servePort := flag.Uint("port", utils.ServePort(), "port to serve on")
	tls := flag.Bool("tls", utils.TLSEnabled(), "serve SSL?")
	certFile := flag.String("cert-file", utils.CertFile(), "path to cert file")
	keyFile := flag.String("key-file", utils.KeyFile(), "SSL Key file path")

	flag.Parse()

	buildRouter()

	if *tls {
		logger.Info("STONKS", "serving https")
		go http.ListenAndServeTLS(fmt.Sprintf(":%d", *servePort), *certFile, *keyFile, router)
	} else {
		logger.Info("STONKS", "serving http")
		go http.ListenAndServe(fmt.Sprintf(":%d", *servePort), router)
	}
	go stocktwits.SuggestedStream(stocktwitsChan, time.Minute)

	go func() {
		for {
			msgs := <-stocktwitsChan
			ws.Broadcast(msgs)
			twitsMessages.SetBulk(msgs)
		}
	}()

	go func() {
		for {
			discussion, waitTime, err := utils.GetBoard("wallstreetbets")
			if err != nil {
				logger.Error("reddit", "get discussion", err.Error())
			}
			redditMessages.SetBulk(discussion)
			time.Sleep(waitTime)
			logger.Info("reddit", "sleeping...")
			time.Sleep(30 * time.Second)
		}
	}()

	logger.Info("START", fmt.Sprintf("listening on port %d", *servePort))
	utils.CloseHandler()

}
