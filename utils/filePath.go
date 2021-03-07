package utils

import (
	"log"
	"os"
	"strings"
)

func subStr(s string, pos int, length int) string {
	r := []rune(s)
	l := pos + length
	if l > len(r) {
		l = len(r)
	}
	// fmt.Println("打印：", s[pos:l])
	return string(s[pos:l])
}
func GetParentPath() string {
	dir := GetCurPath()
	return subStr(dir, 0, strings.LastIndex(dir, "\\"))
}

func GetCurPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(dir)
	}
	return dir
}
