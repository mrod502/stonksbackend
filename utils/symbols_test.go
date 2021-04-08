package utils

import (
	"fmt"
	"testing"
)

func TestRegex(t *testing.T) {

	str := "LFG $GME to the #MOON"

	fmt.Println(GetSymbols(str))
}
