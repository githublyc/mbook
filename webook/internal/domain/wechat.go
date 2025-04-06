package domain

type WeChatInfo struct {
	// 授权用户唯一标识
	OpenId string `json:"openid"`
	// 当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
	UnionId string `json:"unionid"`
}
