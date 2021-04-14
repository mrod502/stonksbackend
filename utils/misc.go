package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

//CloseHandler - wait for ctrl c to close
func CloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("ctrl+c pressed, exiting")
}

//EnableCORS - enable cross-origin requests
func EnableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "privatekey")
	//(*w).Header().Set("access-control-allow-origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST,HEAD,DELETE,PUT")

}

func GetInt64HeaderVal(h http.Header, key string) (v int64, err error) {
	val := h.Get(key)
	if val == "" {
		return 0, errors.New("key not found")
	}
	return strconv.ParseInt(val, 10, 64)
}

//BrowserRequest -- pretend to be a browser so we can get comments
func BrowserRequest(url string, authority ...string) (b []byte, rh http.Header, err error) {

	r, _ := http.NewRequest("GET", url, nil)

	r.Header.Set("upgrade-insecure-requests", "1")
	r.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36")
	r.Header.Set("accept-language", "en-US,en;q=0.9,hr-HR;q=0.8,hr;q=0.7,ru-RU;q=0.6,ru;q=0.5")
	r.Header.Set("scheme", "https")
	r.Header.Set("authority", "www.reddit.com")
	if len(authority) == 1 {
		r.Header.Set("authority", authority[0])
	} else {
		r.Header.Set("authority", "www.reddit.com")
	}
	cli := http.DefaultClient

	resp, err := cli.Do(r)
	if err != nil {
		return
	}
	rh = resp.Header

	b, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return
}
