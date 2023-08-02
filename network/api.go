package network

// type ApiData struct {
// 	Code int64     `json:"code"`
// 	Data []Replies `json:"data"`
// }

// type Replies struct {
// 	Member struct {
// 		Mid   string `json:"mid"`
// 		Uname string `json:"uname"`
// 	} `json:"member"`
// 	Content struct {
// 		Message string `json:"message"`
// 	} `json:"content"`
// }

// // 返回最近回复
// func GetReplies() []Replies {
// 	BaseURL := "https://aliyun.nana7mi.link/comment.get_comments(%v,comment.CommentResourceType.DYNAMIC:parse,1:int).replies"

// 	resp, err := http.Get(fmt.Sprintf(BaseURL, cfg.Oid))
// 	if !printErr(err) {
// 		return nil
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	defer resp.Body.Close()
// 	if !printErr(err) {
// 		return nil
// 	}

// 	var Api ApiData
// 	err = json.Unmarshal(body, &Api)
// 	if !printErr(err) {
// 		return nil
// 	}

// 	if Api.Code != 0 {
// 		return nil
// 	}

// 	return Api.Data
// }
