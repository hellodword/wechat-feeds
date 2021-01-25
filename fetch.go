package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func fetch(ctx context.Context, u string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}

type Detail struct {
	//Name        string `json:"name"`
	Bizid string `json:"bizid"`
	//Description string `json:"description"`
	HeadImg string `json:"head_img"`
	//LastUpdate  int64  `json:"last_update"`
}

func fetchDetails(ctx context.Context) ([]Detail, error) {
	body, err := fetch(ctx, "https://raw.githubusercontent.com/hellodword/wechat-feeds/feeds/details.json")
	if err != nil {
		return nil, err
	}

	var details []Detail
	err = json.Unmarshal(body, &details)
	return details, err
}
