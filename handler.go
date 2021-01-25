package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func writeFavicon(w http.ResponseWriter, name string, bApple bool) {
	var favicon interface{}
	var ok bool
	if bApple {
		favicon, ok = mapFaviconApple.Load(name)
	} else {
		favicon, ok = mapFaviconNormal.Load(name)
	}

	if !ok {
		writeError(w)
		return
	}

	img := favicon.([]byte) // formatImage(favicon.([]byte), bApple)
	// https://stackoverflow.com/questions/56131723/in-http-response-for-a-chunked-data-how-to-set-content-length
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(img)))
	//w.WriteHeader(http.StatusOK)

	_, err := w.Write(img)
	if err != nil {
		log.Println("fmt.Fprintf", err)
	}

}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := fmt.Fprint(w, "bad request")

	if err != nil {
		log.Println("fmt.Fprintf", err)
	}
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	w.Header().Add("Access-Control-Allow-Methods", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	bApple := r.URL.Query().Get("apple") != ""

	bizid := formatBizID(strings.Split(r.URL.Query().Get("host"), ".")[0])
	if bizid == "" || bizid == DefaultFavicon {
		writeFavicon(w, DefaultFavicon, bApple)
		return
	}

	var err error
	_, ok := mapFaviconNormal.Load(bizid)
	if ok {
		writeFavicon(w, bizid, bApple)
		return
	}

	headImg, ok := mapHeadImg.Load(bizid)
	if !ok {
		_ = syncHeadImgs(r.Context())
		headImg, ok = mapHeadImg.Load(bizid)
	}

	if !ok {
		writeError(w)
		return
	}

	img, err := fetch(r.Context(), headImg.(string))
	if err != nil {
		writeError(w)
		return
	}

	mapFaviconNormal.Store(bizid, formatAndResize(img, SizeNormal))
	mapFaviconApple.Store(bizid,  formatAndResize(img, SizeApple))

	writeFavicon(w, bizid, bApple)
}
