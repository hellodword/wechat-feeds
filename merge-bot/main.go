package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hellodword/wechat-feeds/common"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/google/go-github/v33/github"
)

const (
	Owner = "hellodword"
	Repo  = "wechat-feeds"
	Base  = "main"

	PerPage = 100

	RunEvent  = "pull_request"
	RunStatus = "in_progress"

	PRState = "open"

	WorkFlow = "merge-bot"

	ServiceMax = 20000
	ServicePer = 32
)

/*

 */

func main() {

	ctx := context.Background()

	clientWithToken, client := common.MakeClients(ctx, os.Getenv("GITHUB_ACCESS_TOKEN"))
	_, _ = clientWithToken, client
	client = clientWithToken // for private test

	runNumber := common.GetIntFromEnv("GITHUB_RUN_NUMBER")
	_ = runNumber

	createdAt := time.Now()

	/*
		1. as a mutex, actions 的 schedule 不是很准时, 只好多运行一些
		2. github actions 故障时可能 hang 很久, 所以用 createdAt 判断
	*/
	for time.Now().Sub(createdAt) < time.Hour*1 {
		var b bool
		fmt.Println("getting begin time", runNumber)
		createdAt, b = getBeginTime(ctx, client, runNumber)
		fmt.Println("begin", createdAt, time.Now().Sub(createdAt))
		if !b {
			fmt.Println("not the earliest")
			time.Sleep(time.Second * 60)
			continue
		}
		break
	}

	for time.Now().Sub(createdAt) < time.Hour*1 {
		fmt.Println("getPR")
		pr := getPR(ctx, clientWithToken, client)
		if pr == nil {
			fmt.Println("nothing to do")
			time.Sleep(time.Second * 30)
			continue
		}

		fmt.Println("MergeableState", pr.GetMergeableState(), "Mergeable", pr.GetMergeable())

		fmt.Printf("checkPRDetails #%d\n", pr.GetNumber())
		if !checkPRDetails(ctx, clientWithToken, client, pr) {
			fmt.Printf("checkPRDetails #%d failed\n", pr.GetNumber())
			time.Sleep(time.Second * 30)
			continue
		}

		fmt.Printf("merging #%d\n", pr.GetNumber())
		fmt.Println("MergeableState", pr.GetMergeableState(), "Mergeable", pr.GetMergeable())
		_, _, err := clientWithToken.PullRequests.Merge(ctx, Owner, Repo, pr.GetNumber(), "",
			&github.PullRequestOptions{
				MergeMethod: "rebase",
			})
		if err != nil {
			fmt.Printf("merge #%d failed\n", pr.GetNumber())
			closePR(ctx, clientWithToken, pr, common.LabelInvalid,
				fmt.Sprintf("合并出错, 多半是已经冲突了, 重来吧: %s", err.Error()))
		} else {
			fmt.Printf("merge #%d succeeded\n", pr.GetNumber())
			_, _, _ = clientWithToken.Issues.CreateComment(ctx, Owner, Repo, pr.GetNumber(),
				&github.IssueComment{
					Body: github.String("恭喜! 已被合并"),
				})
		}

		time.Sleep(time.Second * 30)
	}

	os.Exit(0)
}

func checkCommitMessage(s string) bool {
	return regexp.MustCompile(`^新增:( [^\s\n\r]+)+$`).MatchString(s)
}

// 150 lines
func checkPRDetails(ctx context.Context, clientWithToken, client *github.Client, pr *github.PullRequest) bool {
	cs, _, err := client.PullRequests.ListCommits(ctx, Owner, Repo, pr.GetNumber(),
		&github.ListOptions{
			//Page:    0,
			PerPage: 2,
		})
	if err != nil {
		return false
	}

	if len(cs) != 1 {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("这个 pr [有 %d 个 commits](https://github.com/hellodword/wechat-feeds/pull/%d/commits)，请确保只有一个 commit，你可以关闭这个 pr 重新提一个。",
				len(cs),
				pr.GetNumber()))
		return false
	}

	fmt.Println(cs[0].GetCommit().GetMessage())
	if pr.GetTitle() != cs[0].GetCommit().GetMessage() {
		fmt.Println("不一致")
	}

	if !checkCommitMessage(strings.Split(cs[0].GetCommit().GetMessage(), "\n")[0]) {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("[提交信息](https://github.com/hellodword/wechat-feeds/pull/%d/commits) 不符合格式，仔细阅读第三步。\n为了方便自动化，所以需要定一个格式。",
				pr.GetNumber()))
		return false
	}

	fs, _, err := client.PullRequests.ListFiles(ctx, Owner, Repo, pr.GetNumber(),
		&github.ListOptions{
			//Page:    0,
			PerPage: 2,
		})

	if err != nil {
		return false
	}

	if len(fs) != 1 || fs[0].GetFilename() != "list.csv" {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("这个 pr [修改了 %d 个文件](https://github.com/hellodword/wechat-feeds/pull/%d/files)，请确保只修改了 list.csv",
				len(fs),
				pr.GetNumber()))
		return false
	}

	fmt.Println(fs[0].GetRawURL())

	// // for private test
	// a, _, _, _ := clientWithToken.Repositories.GetContents(ctx, Owner, Repo, "list.csv", &github.RepositoryContentGetOptions{Ref: cs[0].GetSHA()})
	// b, _ := a.GetContent()
	// newBody := []byte(b)

	newBody := common.Fetch(fs[0].GetRawURL())
	if !common.WithUTF8Bom(newBody) {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			"list.csv 必须是 UTF8-BOM，请不要修改格式")
		return false
	}

	var newBis []*common.BizInfo
	err = gocsv.Unmarshal(bytes.NewReader(common.TrimUTF8Bom(newBody)), &newBis)
	if err != nil {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("list.csv 解析失败: %s", err.Error()))
		return false
	}

	var oldBody []byte
	var oldBis []*common.BizInfo
	for {
		oldBody = common.Fetch("https://github.com/hellodword/wechat-feeds/raw/main/list.csv") // fetch(fs[0].GetRawURL())
		if !common.WithUTF8Bom(oldBody) {
			fmt.Println("???", string(oldBody))
			continue
		}

		err = gocsv.Unmarshal(bytes.NewReader(common.TrimUTF8Bom(oldBody)), &oldBis)
		if err != nil {
			fmt.Println("???", err, string(oldBody))
			continue
		}
		break
	}

	if len(oldBis) >= ServiceMax {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("已超过本服务限额 %d 个公众号，暂不接受添加新的公众号: %d",
				ServiceMax,
				len(oldBis)))
		return false
	}

	fmt.Println(len(newBis), len(oldBis))
	if len(newBis)-len(oldBis) <= 0 {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("原条目 %d，新条目 %d", len(oldBis), len(newBis)))
		return false
	}

	if len(newBis)-len(oldBis) > ServicePer {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("条目变化 %d 多于 %d，每次请不要添加多于 %d 个公众号",
				len(newBis)-len(oldBis),
				ServicePer,
				ServicePer))
		return false
	}

	// 先检查是否有删除
	var deleted []string
LABEL1:
	for i := range oldBis {
		for j := range newBis {
			if oldBis[i].BizID == newBis[j].BizID {
				continue LABEL1
			}
		}
		deleted = append(deleted, oldBis[i].BizID)
	}

	if len(deleted) != 0 {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("自助添加不支持删除公众号，如果确定需要删除，请 `@hellodword` 等待手动处理"))
		return false
	}

	// 检查有无重复
	var duplicated []string
	m := map[string]int8{}
	for i := range newBis {

		if !common.CheckBizIDSimple(newBis[i].BizID) {
			closePR(ctx, clientWithToken, pr, common.LabelInvalid,
				fmt.Sprintf("解析出无效的 bizid   %s", newBis[i].BizID))
			return false
		}

		_, ok := m[newBis[i].BizID]
		if ok {
			duplicated = append(duplicated, newBis[i].BizID)
		} else {
			m[newBis[i].BizID] = 0
		}
	}

	if len(duplicated) != 0 {
		closePR(ctx, clientWithToken, pr, common.LabelInvalid,
			fmt.Sprintf("以下 bizid 重复，请重新提交\n\n%s", strings.Join(duplicated, "\n")))
		return false
	}

	return true
}

func wrapCommentWithHeader(s string) string {
	return fmt.Sprintf("**下方消息由 merge-bot 自动发送, 请仔细阅读**\n\n**错误提示**: \n%s\n\n如果你 **很确定以上错误提示不是你的问题**, 可以 `@hellodword` 呼叫作者", s)
}

func closePR(ctx context.Context, clientWithToken *github.Client, pr *github.PullRequest, label common.Label, comment string) {
	fmt.Printf("closing pr #%d %s %s\n", pr.GetNumber(), pr.GetTitle(), label)

	_, _, err := clientWithToken.Issues.AddLabelsToIssue(ctx, Owner, Repo, pr.GetNumber(), []string{string(label)})
	if err != nil {
		fmt.Println("AddLabelsToIssue")
	}

	pr.State = github.String(string(common.StateClosed))
	_, _, err = clientWithToken.PullRequests.Edit(ctx, Owner, Repo, pr.GetNumber(), pr)
	if err != nil {
		fmt.Println("Edit")
	}

	_, _, err = clientWithToken.Issues.CreateComment(ctx, Owner, Repo, pr.GetNumber(),
		&github.IssueComment{
			Body: github.String(wrapCommentWithHeader(comment)),
		})
	if err != nil {
		fmt.Println("CreateComment")
	}
}

func getPR(ctx context.Context, clientWithToken, client *github.Client) *github.PullRequest {
	prs, _, err := client.PullRequests.List(ctx, Owner, Repo,
		&github.PullRequestListOptions{
			State: PRState,
			//Head:        "",
			Base:      Base,
			Sort:      "created",
			Direction: "desc",
			ListOptions: github.ListOptions{
				//Page:    0,
				PerPage: PerPage,
			},
		})

	if err != nil {
		fmt.Println("PullRequests.List")
		return nil
	}

	for i := range prs {
		fmt.Println(prs[i].Title, len(prs[i].Labels), prs[i].MergedAt != nil, prs[i].GetMergeable(), prs[i].GetMergeableState())
		if len(prs[i].Labels) > 0 {
			closePR(ctx, clientWithToken, prs[i], common.LabelUB,
				"不支持此操作，请不要再尝试 reopen 这个 pr")
			continue
		}
		if prs[i].MergedAt != nil {
			closePR(ctx, clientWithToken, prs[i], common.LabelUB,
				"不支持此操作，请不要再尝试 reopen 这个 pr")
			continue
		}
		return prs[i]
	}

	return nil
}

func getBeginTime(ctx context.Context, client *github.Client, num int) (time.Time, bool) {
	wrs, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, Owner, Repo,
		&github.ListWorkflowRunsOptions{
			Actor: Owner,
			//Branch:      "",
			//Event:  RunEvent,
			Status: RunStatus,
			ListOptions: github.ListOptions{
				//Page:    0,
				PerPage: PerPage,
			},
		})

	if err != nil {
		return time.Time{}, false
	}

	if wrs == nil || len(wrs.WorkflowRuns) == 0 {
		panic("no runs")
		return time.Time{}, false
	}

	var nums []int
	for i := range wrs.WorkflowRuns {
		nums = append(nums, wrs.WorkflowRuns[i].GetRunNumber())
	}

	if nums == nil { // go lint
		panic("")
	}

	sort.Ints(nums)
	fmt.Println(num, nums)
	for i := range wrs.WorkflowRuns {
		if num == wrs.WorkflowRuns[i].GetRunNumber() &&
			WorkFlow == wrs.WorkflowRuns[i].GetName() { // 小心: GetName 是 go-github 尚未发布的 API
			return wrs.WorkflowRuns[i].GetCreatedAt().Time, num == nums[0]
		}
	}

	panic("can not find this run num")

}
