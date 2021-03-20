package main

import (
	"flag"
	"fmt"

	"github.com/mrod502/stonksbackend/utils"
)

func main() {
	configFilePath := flag.String("config-file-path", "settings.json", "a string")

	flag.Parse()

	utils.ReadConfig(*configFilePath)

	dataSources := utils.DataSources()

	fmt.Println(dataSources)

	return
}
