package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
)

func main() {
	bis, e := getList()
	if e != nil {
		fmt.Println("get list error", e)
		os.Exit(0)
	}
	bds, e := getDetails()
	if e != nil {
		fmt.Println("get details error", e)
		os.Exit(0)
	}

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
		fmt.Println(bs)
	} else {
		fmt.Println("一切正常")
	}
	os.Exit(0)
}

type bizInfo struct {
	Name        string `csv:"name"`
	BizID       string `csv:"bizid"`
	Description string `csv:"description"`
}

func getList() ([]*bizInfo, error) {

	r, e := http.Get("https://github.com/hellodword/wechat-feeds/raw/main/list.csv")
	if e != nil {
		return nil, e
	}

	defer r.Body.Close()

	bis := []*bizInfo{}
	e = gocsv.Unmarshal(r.Body, &bis)
	if e != nil {
		return nil, e
	}

	return bis, nil
}

type bizDetail struct {
	Name  string `csv:"name" json:"name"`
	BizID string `csv:"bizid" json:"bizid"`
}

func getDetails() ([]*bizDetail, error) {
	r, e := http.Get("https://github.com/hellodword/wechat-feeds/raw/feeds/details.json")
	if e != nil {
		return nil, e
	}

	defer r.Body.Close()

	bds := []*bizDetail{}
	e = json.NewDecoder(r.Body).Decode(&bds)
	if e != nil {
		return nil, e
	}

	return bds, nil
}
