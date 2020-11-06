package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
)

func main() {
	bis := getList()
	bds := getDetails()

	bs := ""

LABEL:
	for i := range bis {
		for j := range bds {
			if bis[i].BizID == bds[j].BizID {
				continue LABEL
			}
		}

		bs += bis[i].BizID
		bs += "\n"
	}

	if bs != "" {
		fmt.Fprintf(os.Stderr, "%s", bs)
		os.Exit(1) // crash to email me
	}
}

type bizInfo struct {
	Name        string `csv:"name"`
	BizID       string `csv:"bizid"`
	Description string `csv:"description"`
}

func getList() []*bizInfo {

	r, e := http.Get("https://github.com/hellodword/wechat-feeds/raw/main/list.csv")
	if e != nil {
		panic(e)
	}

	defer r.Body.Close()

	bis := []*bizInfo{}
	e = gocsv.Unmarshal(r.Body, &bis)
	if e != nil {
		panic(e)
	}

	return bis
}

type bizDetail struct {
	Name  string `csv:"name" json:"name"`
	BizID string `csv:"bizid" json:"bizid"`
}

func getDetails() []*bizDetail {
	r, e := http.Get("https://github.com/hellodword/wechat-feeds/raw/feeds/details.json")
	if e != nil {
		panic(e)
	}

	defer r.Body.Close()

	bds := []*bizDetail{}
	e = json.NewDecoder(r.Body).Decode(&bds)
	if e != nil {
		panic(e)
	}

	return bds
}
