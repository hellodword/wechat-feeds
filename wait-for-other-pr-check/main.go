package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {

	for {

		if ready() {
			fmt.Println("没有其它任务了，一分钟后放行")
			time.Sleep(time.Second * 60)
			os.Exit(0)
		}

		time.Sleep(time.Second * 60)
	}

}

func ready() bool {
	res, err := http.Get("https://api.github.com/repos/hellodword/wechat-feeds/actions/runs?event=pull_request&status=in_progress&per_page=100")
	if err != nil {
		// panic(err)
		fmt.Println(err)
		return false
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// panic(err)
		fmt.Println(err)
		return false
	}

	type Data struct {
		TotalCount int64 `json:"total_count"`
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		// panic(err)
		fmt.Println(err)
		return false
	}

	return data.TotalCount <= 1
}
