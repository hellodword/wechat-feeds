package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {

	GITHUB_RUN_NUMBER := os.Getenv("GITHUB_RUN_NUMBER")
	if GITHUB_RUN_NUMBER == "" {
		panic("no GITHUB_RUN_NUMBER")
	}

	fmt.Println("GITHUB_RUN_NUMBER", GITHUB_RUN_NUMBER)

	num, err := strconv.Atoi(GITHUB_RUN_NUMBER)
	if err != nil {
		panic(err)
	}

	count := 0
	for {

		if ready(num) {
			if count == 0 {
				fmt.Println("没有其它任务了，放行")
			} else {
				fmt.Println("没有其它任务了，一分钟后放行")
				time.Sleep(time.Second * 60)
			}
			os.Exit(0)
		}
		count++
		fmt.Println("先来后到，候着吧")
		time.Sleep(time.Second * 60)
	}

}

func ready(num int) bool {
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
		TotalCount   int64 `json:"total_count"`
		WorkflowRuns []struct {
			ID        int64 `json:"id"`
			RunNumber int64 `json:"run_number"`
		} `json:"workflow_runs"`
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		// panic(err)
		fmt.Println(err)
		return false
	}

	if data.TotalCount <= 1 {
		return true
	}

	var nums []int
	for i := range data.WorkflowRuns {
		nums = append(nums, int(data.WorkflowRuns[i].RunNumber))
	}

	if len(nums) <= 1 {
		return true
	}

	sort.Ints(nums)

	return num == nums[0]
}
