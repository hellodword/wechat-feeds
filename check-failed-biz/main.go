package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v33/github"
	"github.com/hellodword/wechat-feeds/common"
	"os"

	"github.com/gocarina/gocsv"
)

const (
	Owner = "hellodword"
	Repo  = "wechat-feeds"

	IssueID = 608
)

func main() {

	ctx := context.Background()

	//var event github.IssueEvent
	//common.ParseGithubEvent(&event)

	clientWithToken, client := common.MakeClients(ctx, os.Getenv("GITHUB_ACCESS_TOKEN"))
	_, _ = clientWithToken, client
	client = clientWithToken // for private test

	_, _, _ = clientWithToken.Issues.AddLabelsToIssue(ctx, Owner, Repo,
		IssueID,
		[]string{string(common.LabelCheck)})

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

	var body string
	if bs != "" {
		bs = "以下 bizid 可能有问题或尚未同步：\n" + bs
		//fmt.Println(bs)
		body = bs
	} else {
		//fmt.Println("一切正常")
		body = "一切正常"
	}

	fmt.Println(body)

	_, _, err := clientWithToken.Issues.CreateComment(ctx, Owner, Repo,
		IssueID,
		&github.IssueComment{
			Body: github.String(body),
		})
	if err != nil {
		panic("Issues.CreateComment") // token privacy
	}

	os.Exit(0)
}

func getList() []*common.BizInfo {

	body := common.Fetch("https://github.com/hellodword/wechat-feeds/raw/main/list.csv")
	if !common.WithUTF8Bom(body) {
		panic("list.csv not utf8 bom")
	}

	var bis []*common.BizInfo
	err := gocsv.Unmarshal(bytes.NewReader(common.TrimUTF8Bom(body)), &bis)
	if err != nil {
		panic(err)
	}

	return bis
}

func getDetails() []*common.BizDetail {

	body := common.Fetch("https://github.com/hellodword/wechat-feeds/raw/feeds/details.json")

	var bds []*common.BizDetail
	err := json.Unmarshal(body, &bds)
	if err != nil {
		panic(err)
	}

	return bds
}
