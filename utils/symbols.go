package utils

import "regexp"

var (
	ptnTag = regexp.MustCompile(`[$#]([A-Z.]{1,8})`)
)

//GetSymbols get all symbols
func GetSymbols(s string) (m []string) {
	res := ptnTag.FindAllStringSubmatch(s, -1)
	for _, v := range res {
		m = append(m, v[1])
	}
	return
}
