package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

func Vardump(a ...interface{}) {
	fmt.Println(spew.Sdump(a))
}

func Hex2String(h string) string {
	b, _ := hex.DecodeString(h[2:])
	return string(b)
}
