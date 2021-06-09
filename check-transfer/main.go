package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v33/github"
	"github.com/hellodword/wechat-feeds/common"
	"github.com/mmcdole/gofeed"
	"math/rand"
	"net/url"
	"os"
	"time"
)

const (
	Owner = "hellodword"
	Repo  = "wechat-feeds"

	IssueID = 2387
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	details := getDetails()
	if len(details) == 0 {
		os.Exit(0)
	}

	ctx := context.Background()
	clientWithToken, client := common.MakeClients(ctx, os.Getenv("GITHUB_ACCESS_TOKEN"))
	_, _ = clientWithToken, client
	client = clientWithToken // for private test

	_, _, _ = clientWithToken.Issues.AddLabelsToIssue(ctx, Owner, Repo,
		IssueID,
		[]string{string(common.LabelCheck)})

	buf := bytes.NewBuffer(nil)
	//	buf.WriteString(`| name | bizid | new | reason |
	////| ---  | ----- | --- | ------ |`)
	for i := range details {

		newBiz, reason := checkTransfer(details[i].BizID)
		if newBiz == "" && reason == "" {
			continue
		}
		if newBiz == details[i].BizID {
			continue
		}

		buf.WriteString("\n")
		buf.WriteString("| ")
		buf.WriteString(details[i].Name)
		buf.WriteString(" | ")
		buf.WriteString(fmt.Sprintf(`[%s](https://github.com/hellodword/wechat-feeds/raw/feeds/%s.xml)`,
			details[i].BizID, details[i].BizID))
		buf.WriteString(" | ")
		buf.WriteString(newBiz)
		buf.WriteString(" | ")
		buf.WriteString(reason)
		buf.WriteString(" | ")

	}

	fmt.Println(buf.String())
	if buf.String() != "" {
		_, _, err := clientWithToken.Issues.CreateComment(ctx, Owner, Repo,
			IssueID,
			&github.IssueComment{
				Body: github.String(`| name | bizid | new | reason |
| ---  | ----- | --- | ------ |` + buf.String()),
			})
		if err != nil {
			panic("Issues.CreateComment") // token privacy
		}

	}

	os.Exit(0)
}

func checkTransfer(bizid string) (newBiz, reason string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(common.Fetch(fmt.Sprintf("https://github.com/hellodword/wechat-feeds/raw/feeds/%s.xml", url.QueryEscape(bizid)))))
	if err != nil || feed == nil || len(feed.Items) == 0 {
		reason = "解析失败" // 随便处理一下
		return
	}

	i := rand.Intn(len(feed.Items)) // 原文也失效 https://mp.weixin.qq.com/s/6kgFxZ6nm9dLRZV66-9RXA

	articleInfo, _ := common.FetchWX(feed.Items[i].Link)

	newBiz = articleInfo.BizID
	reason = articleInfo.FailReason

	if newBiz == "" && articleInfo.TransferLink != "" { // 原文有问题，直接从链接取 bizid
		// http://mp.weixin.qq.com/s?__biz=Mzk0MDIwNTQxNw==&amp;mid=2247505407&amp;idx=1&amp;sn=62616bd14bfe1eb8f570442ffddefcd0#rd
		newBiz = common.MatchBizID(articleInfo.TransferLink)
	}

	return
}

func getDetails() []*common.BizDetail {

	body := common.Fetch("https://github.com/hellodword/wechat-feeds/raw/feeds/details.json")

	var bds []*common.BizDetail
	err := json.Unmarshal(body, &bds)
	if err != nil {
		panic(err)
	}

	// pick out details without head_img

	var ret []*common.BizDetail
	for i := range bds {
		if bds[i].HeadIMG == "" {
			ret = append(ret, bds[i])
		}
	}

	return ret
}
