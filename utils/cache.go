package utils

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mrod502/logger"
)

// map-related errors
var (
	ErrKeyNotFound    = errors.New("key not found or value at key is nil")
	ErrExpiryTooShort = errors.New("expiry time is too short")
	ErrKeyExists      = errors.New("key already exists")
)

type sendError struct {
	k   string
	msg error
}

type value struct {
	t       time.Time
	v       *websocket.Conn
	m       chan []byte
	errChan chan sendError
	k       string
}

func (v value) sender() {
	for {
		v.v.SetWriteDeadline(time.Now().Add(3 * time.Second))
		if err := v.v.WriteMessage(websocket.TextMessage, <-v.m); err != nil {
			v.errChan <- sendError{k: v.k, msg: err}
			return
		}
	}
}

//WebsocketMap - handles broadcasting to multiple websocket conns
type WebsocketMap struct {
	v         map[string]*value
	l         *sync.RWMutex
	expiry    time.Duration
	closeChan chan *websocket.Conn
	errChan   chan sendError
}

//Get - get the underlying weebsocket conn at a key
func (w WebsocketMap) Get(k string) (*websocket.Conn, error) {
	w.l.RLock()
	defer w.l.RUnlock()
	if v, ok := w.v[k]; v != nil && ok {
		return v.v, nil
	}
	return nil, ErrKeyNotFound
}

//Set - initially set a value at a key
func (w WebsocketMap) Set(k string, v *websocket.Conn) (err error) {
	w.l.Lock()
	defer w.l.Unlock()
	if w.v[k] != nil {
		return ErrKeyExists
	}
	w.v[k] = &value{v: v,
		t:       time.Now(),
		m:       make(chan []byte, 128),
		errChan: w.errChan,
	}

	return nil
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
		if k.msg != nil {
			logger.Warn("DELETER", k.msg.Error(), "- closing conn")
		}
		w.l.Lock()
		w.closeChan <- w.v[k.k].v
		delete(w.v, k.k)
		w.l.Unlock()
	}
}

//NewWebsocketMap - initialize a websocket map
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
		errChan:   make(chan sendError, 256),
	}

	if useJanitor {
		go w.janitor()
	}

	go w.closer()

	return w, nil
}

//Broadcast - JSON marshal an object and broadcast to all conns
func (w WebsocketMap) Broadcast(v interface{}) error {
	var b []byte
	var err error
	if b, err = json.Marshal(v); err != nil {
		return err
	}
	w.l.Lock()
	for _, v := range w.v {
		logger.Info("Broadcast", "broadcasting", string(b), "to", v.v.RemoteAddr().String())
		v.m <- b
	}
	w.l.Unlock()

	return nil

}
