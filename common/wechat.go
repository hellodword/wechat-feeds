package common

import (
	"encoding/base64"
	"strconv"
)

func CheckBizIDSimple(bizid string) bool {
	if bizid == "" {
		return false
	}
	b, e := base64.StdEncoding.DecodeString(bizid)
	if e != nil {
		return false
	}
	if base64.StdEncoding.EncodeToString(b) != bizid {
		return false
	}
	s := string(b)
	if s == "" {
		return false
	}

	i, e := strconv.Atoi(s)
	if e != nil {
		return false
	}
	if i <= 0 {
		return false
	}
	return strconv.Itoa(i) == s
}
