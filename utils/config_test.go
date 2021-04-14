package utils

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestConfig(t *testing.T) {
	home, _ := os.UserHomeDir()
	ReadConfig(path.Join(home, "stonksbackend.json"))
	fmt.Printf("%+v", config)
}
