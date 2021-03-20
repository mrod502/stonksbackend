package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/mrod502/stonksbackend/utils"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configFilePath := flag.String("config-file-path", path.Join(home, "settings.json"), "a string")

	flag.Parse()

	utils.ReadConfig(*configFilePath)

	dataSources := utils.DataSources()

	fmt.Println(dataSources)

	return
}
