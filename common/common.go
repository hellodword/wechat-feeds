package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func GetIntFromEnv(key string) int {
	s := os.Getenv(key)
	if s == "" {
		panic(fmt.Errorf("no %s", key))
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return num
}

func Fetch(u string) []byte {
	res, err := http.Get(u)
	if err != nil {
		return nil
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := ioutil.ReadAll(res.Body)
	return body
}

func WithUTF8Bom(body []byte) bool {
	return len(body) > 3 && body[0] == 0xef && body[1] == 0xbb && body[2] == 0xbf
}

func TrimUTF8Bom(body []byte) []byte {
	return body[3:]
}
