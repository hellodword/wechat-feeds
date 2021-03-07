package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v33/github"
	"github.com/hellodword/wechat-feeds/common"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	Owner = "hellodword"
	Repo  = "wechat-feeds"
)

func main() {

	ctx := context.Background()

	var event github.IssueEvent
	common.ParseGithubEvent(&event)

	clientWithToken, client := common.MakeClients(ctx, os.Getenv("GITHUB_ACCESS_TOKEN"))
	_, _ = clientWithToken, client
	client = clientWithToken // for private test

	issueBody := event.GetIssue().GetBody()

	fmt.Println("body", issueBody)

	if strings.HasPrefix(issueBody, "#skip") {
		os.Exit(0)
	}

	us := regexp.MustCompile(`(?m)^https?://mp\.weixin\.qq\.com/s[^\s\r\n]+`).FindAllString(event.GetIssue().GetTitle()+"\n"+issueBody, -1)

	fmt.Println("match", strings.Join(us, "\n"))

	if len(us) == 0 {
		if issueBody == "" { // 有些看文档不仔细，标题里有链接，处理一下吧
			closeIssue(ctx, clientWithToken,
				event.GetIssue(), common.LabelInvalid,
				"issue 内容不可为空，仔细阅读 [如何获取 bizid?](https://github.com/hellodword/wechat-feeds#%E5%A6%82%E4%BD%95%E8%8E%B7%E5%8F%96-bizid)")
		}
		os.Exit(0)
	}

	done := map[string]int{}
	succ := map[string]WXArticle{}
	fail := map[string]error{}

	for i := 0; i < len(us); i++ {
		fmt.Println(len(us), i, us[i])
		_, ok := done[us[i]]
		if ok {
			continue
		}
		done[us[i]] = 0

		if i != 0 {
			time.Sleep(time.Second * 5)
		}

		article, err := fetchWX(us[i])
		if err == nil {
			succ[article.BizID] = article
		} else {
			fail[us[i]] = err
		}
	}

	fmt.Println("*", len(succ), len(fail))

	body := bytes.NewBufferString("抓取结果\n\n")
	if len(succ) > 0 {
		body.WriteString("以下为抓取成功的\n")
		body.WriteString("```\n")
		for k := range succ {
			body.WriteString(fmt.Sprintf("%s,%s,%s\n",
				succ[k].Name,
				succ[k].BizID,
				`""`)) // ignore description
		}
		body.WriteString("```\n\n")
		body.WriteString("以上只是抓取，自行和 list.csv 对比去重\n\n")
	}

	if len(fail) > 0 {
		body.WriteString("以下为抓取失败的\n")
		body.WriteString("```\n")
		for k := range fail {
			body.WriteString(fmt.Sprintf("%s\n%s\n\n",
				k,
				fail[k].Error()))
		}
		body.WriteString("```\n\n")
	}

	closeIssue(ctx, clientWithToken,
		event.GetIssue(), common.LabelFetch,
		body.String())

	os.Exit(0)
}

type WXArticle struct {
	Name        string
	BizID       string
	Description string
}

func getTransferTargetLink(s string) string {
	r := regexp.MustCompile(`transferTargetLink = '(https?://mp\.weixin\.qq\.com/s[^\s\r\n]+)'`).FindStringSubmatch(s)
	if len(r) < 2 {
		return ""
	} else {
		return r[1]
	}
}

func getBizID(s string) string {

	// https://mp.weixin.qq.com/s/1I33XLA5uK1Iljvn3-XVDg
	// https://mp.weixin.qq.com/s/etTO4fTRwyvSUuh2qJlIaw

	r := regexp.MustCompile(`((var biz = [" =|]*")|(var appuin = [" =|]*")|(__biz=))([a-zA-Z\d/+=]+)`).FindStringSubmatch(s)
	if len(r) == 6 {
		return r[5]
	} else {
		return ""
	}
}

func getName(s string) string {

	// 图文 https://mp.weixin.qq.com/s/g0H8YxjN5kUR3Kx9cPgNlA

	r := regexp.MustCompile(`(var nickname = "([^\n]+)";)|(d\.nick_name = getXmlValue\('nick_name.DATA'\) \|\| '([^\n]+)';)|(<strong class="account_nickname_inner js_go_profile">([^\n]+)</strong>)`).FindStringSubmatch(s)
	if len(r) != 7 {
		return ""
	}

	if r[2] != "" {
		return r[2]
	}
	if r[4] != "" {
		return r[4]
	}
	if r[6] != "" {
		return r[6]
	}

	return ""

}

func fetchWX(u string) (article WXArticle, err error) {
	fmt.Println("link", u)

	body := common.Fetch(u)

	s := string(body)

	link := getTransferTargetLink(s)
	if link != "" {
		fmt.Println("transfer link", link)
		return fetchWX(link)
	}

	article.BizID = getBizID(s)
	if article.BizID == "" {
		err = errors.New("no biz id")
		return
	}
	article.Name = getName(s)
	if article.Name == "" {
		err = errors.New("no name")
		return
	}

	return
}

func closeIssue(ctx context.Context, clientWithToken *github.Client, issue *github.Issue, label common.Label, comment string) {
	fmt.Printf("closing issue #%d %s %s\n", issue.GetNumber(), issue.GetTitle(), label)

	_, _, _ = clientWithToken.Issues.Edit(ctx, Owner, Repo,
		issue.GetNumber(),
		&github.IssueRequest{
			State: github.String(string(common.StateClosed)),
		})
	_, _, _ = clientWithToken.Issues.AddLabelsToIssue(ctx, Owner, Repo,
		issue.GetNumber(),
		[]string{string(label)})
	_, _, _ = clientWithToken.Issues.CreateComment(ctx, Owner, Repo,
		issue.GetNumber(),
		&github.IssueComment{
			Body: github.String(comment),
		})
}
