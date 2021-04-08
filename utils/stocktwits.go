package utils

import (
	"sync"

	stocktwits "github.com/mrod502/stocktwitsgo"
)

type StocktwitsCache struct {
	v map[int]stocktwits.Message
	sync.RWMutex
}

func (c *StocktwitsCache) Get(k int) stocktwits.Message {
	c.RLock()
	defer c.RUnlock()
	return c.v[k]
}

func (c *StocktwitsCache) Set(v stocktwits.Message) {
	c.Lock()
	defer c.Unlock()

	c.v[v.ID] = v
}

func (c *StocktwitsCache) SetBulk(vals []stocktwits.Message) {
	c.Lock()
	defer c.Unlock()
	for _, v := range vals {
		c.v[v.ID] = v
	}
}

func (c *StocktwitsCache) All() (d []stocktwits.Message) {

	d = make([]stocktwits.Message, 0, len(c.v))
	c.RLock()
	defer c.RUnlock()
	for _, v := range c.v {
		d = append(d, v)
	}
	return
}
func NewStocktwitsCache() *StocktwitsCache {
	return &StocktwitsCache{v: make(map[int]stocktwits.Message)}
}
