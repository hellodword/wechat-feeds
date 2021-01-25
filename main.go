package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

const (
	DefaultFavicon = "default"
)

const (
	SizeNormal = 32 //196
	SizeApple  = 57 //180
)

var mapFaviconNormal sync.Map
var mapFaviconApple sync.Map
var mapHeadImg sync.Map

func init() {

	r, err := http.Get("https://wechat.privacyhide.com/favicon.ico")
	if err != nil {
		panic(err)
	}

	if r.StatusCode != http.StatusOK {
		panic(fmt.Errorf("status code %d", r.StatusCode))
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	contentType := http.DetectContentType(b)
	if contentType != "image/x-icon" {
		panic(fmt.Errorf("content type %s", contentType))
	}

	mapFaviconNormal.Store(DefaultFavicon, b)

	err = syncHeadImgs(context.Background())
	if err != nil {
		panic(err)
	}

}

func syncHeadImgs(ctx context.Context) error {
	details, err := fetchDetails(ctx)
	if err != nil {
		return err
	}

	for i := range details {
		if details[i].HeadImg != "" {
			mapHeadImg.Store(formatBizID(details[i].Bizid), formatHeadImg(details[i].HeadImg))
		}
	}
	return nil
}

func formatHeadImg(headImg string) string {
	if strings.HasSuffix(headImg, "/132") {
		headImg = strings.TrimSuffix(headImg, "/132")
		headImg = headImg + "/64"
	}
	return headImg
}

func formatBizID(bizid string) string {
	return strings.ToLower(strings.ReplaceAll(bizid, "=", ""))
}

func main() {

	handler := http.NewServeMux()
	handler.HandleFunc("/favicon.ico", handleFavicon)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	panic(server.ListenAndServe())

}
