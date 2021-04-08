package utils

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

var (
	config     *RuntimeConfig
	configLock *sync.RWMutex
)

func init() {
	config = &RuntimeConfig{DataSources: make(map[string][]string)}
	configLock = new(sync.RWMutex)
}

//RuntimeConfig - configure the runtime
type RuntimeConfig struct {
	DataSources map[string][]string `json:"data-sources"`
	ServePort   int32               `json:"serve-port"`
	CertFile    string              `json:"cert-file"`
	KeyFile     string              `json:"key-file"`
}

//ReadConfig - read file at configPath and load into config
func ReadConfig(configPath string) (err error) {
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	configLock.Lock()
	defer configLock.Unlock()
	return json.Unmarshal(b, config)
}

//DataSources - return datasources
func DataSources() map[string][]string {
	configLock.RLock()
	defer configLock.RUnlock()
	return config.DataSources
}

//ServePort - return serveport
func ServePort() int32 {
	configLock.RLock()
	defer configLock.RUnlock()
	return config.ServePort
}

func CertFile() string {
	configLock.RLock()
	defer configLock.RUnlock()
	return config.CertFile
}

func KeyFile() string {
	configLock.RLock()
	defer configLock.RUnlock()
	return config.KeyFile
}
