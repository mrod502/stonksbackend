package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrKeyNotFound    = errors.New("key not found or value at key is nil")
	ErrExpiryTooShort = errors.New("expiry time is too short")
)

type value struct {
	t       time.Time
	v       *websocket.Conn
	m       chan []byte
	errChan chan string
	k       string
}

func (v value) sender() {
	for {
		if err := v.v.WriteMessage(websocket.TextMessage, <-v.m); err != nil {
			v.errChan <- v.k
		}
	}
}

type WebsocketMap struct {
	v         map[string]*value
	l         *sync.RWMutex
	expiry    time.Duration
	closeChan chan *websocket.Conn
	errChan   chan string
}

func (w WebsocketMap) Get(k string) (*websocket.Conn, error) {
	w.l.RLock()
	defer w.l.RUnlock()
	if v, ok := w.v[k]; v != nil && ok {
		return v.v, nil
	}
	return nil, ErrKeyNotFound
}

func (w WebsocketMap) Set(k string, v *websocket.Conn) {
	w.l.Lock()
	defer w.l.Unlock()
	w.v[k] = &value{v: v, t: time.Now()}
}

func (w WebsocketMap) janitor() {
	for {
		time.Sleep(time.Minute)
		now := time.Now()
		w.l.Lock()
		for k, v := range w.v {
			if v.t.Sub(now.Add(-w.expiry)) < 0 {
				w.closeChan <- v.v
			}
			delete(w.v, k)
		}
		w.l.Unlock()
	}
}

func (w WebsocketMap) closer() {
	for {
		c := <-w.closeChan
		_ = c.Close()
	}
}

func (w WebsocketMap) keyDeleter() {
	for {
		k := <-w.errChan
		w.l.Lock()
		w.closeChan <- w.v[k].v
		delete(w.v, k)
		w.l.Unlock()
	}
}

func NewWebsocketMap(expiry time.Duration, useJanitor bool) (*WebsocketMap, error) {
	if useJanitor {
		if expiry < time.Minute {
			return nil, ErrExpiryTooShort
		}
	}

	w := &WebsocketMap{
		v:         make(map[string]*value),
		l:         new(sync.RWMutex),
		expiry:    expiry,
		closeChan: make(chan *websocket.Conn, 256),
	}

	if useJanitor {
		go w.janitor()
	}

	go w.closer()

	return w, nil
}

func (w WebsocketMap) Broadcast(v interface{}) error {

	return nil

}
