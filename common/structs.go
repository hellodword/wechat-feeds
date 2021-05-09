package common

type BizDetail struct {
	Name    string `csv:"name" json:"name"`
	BizID   string `csv:"bizid" json:"bizid"`
	HeadIMG string `json:"head_img"`
}

type BizInfo struct {
	Name        string `csv:"name"`
	BizID       string `csv:"bizid"`
	Description string `csv:"description"`
}
