package common

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func MatchTransferTargetLink(s string) string {
	r := regexp.MustCompile(`transferTargetLink = '(https?://mp\.weixin\.qq\.com/s[^\s\r\n]+)'`).FindStringSubmatch(s)
	if len(r) < 2 {
		return ""
	} else {
		return r[1]
	}
}

func MatchBizID(s string) string {

	// https://mp.weixin.qq.com/s/1I33XLA5uK1Iljvn3-XVDg
	// https://mp.weixin.qq.com/s/etTO4fTRwyvSUuh2qJlIaw

	r := regexp.MustCompile(`((var biz = [" =|]*")|(var appuin = [" =|]*")|(__biz=))([a-zA-Z\d/+=]+)`).FindStringSubmatch(s)
	if len(r) == 6 {
		return r[5]
	} else {
		return ""
	}
}

func MatchName(s string) string {

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

type WXArticle struct {
	Name        string
	BizID       string
	Description string

	FailReason   string
	TransferLink string
}

func FetchWX(u string) (article WXArticle, err error) {
	fmt.Println("link", u)

	body := Fetch(u)

	s := string(body)

	if article.FailReason == "" {
		if strings.Index(s, `此帐号已被屏蔽, 内容无法查看`) != -1 {
			article.FailReason = `此帐号已被屏蔽, 内容无法查看`
			err = errors.New(article.FailReason)
			return
		} else if strings.Index(s, `此帐号已自主注销，内容无法查看`) != -1 {
			article.FailReason = `此帐号已自主注销，内容无法查看`
			err = errors.New(article.FailReason)
			return
		} else if strings.Index(s, `原帐号迁移时未将文章素材同步至新帐号，该链接已不可访问`) != -1 {
			article.FailReason = `原帐号迁移时未将文章素材同步至新帐号，该链接已不可访问`
			err = errors.New(article.FailReason)
			return
		} else if strings.Index(s, `该公众号已迁移`) != -1 {
			article.FailReason = `该公众号已迁移`
		}

	}

	link := MatchTransferTargetLink(s)
	if link != "" {
		fmt.Println("transfer link", link)
		newArticle, err := FetchWX(link)
		newArticle.TransferLink = link
		if article.FailReason != "" && newArticle.FailReason == "" {
			newArticle.FailReason = article.FailReason
		}
		return newArticle, err
	}

	article.BizID = MatchBizID(s)
	if article.BizID == "" {
		err = errors.New("no biz id")
		return
	}
	article.Name = MatchName(s)
	if article.Name == "" {
		err = errors.New("no name")
		return
	}

	return
}
