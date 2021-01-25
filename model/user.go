package model

type UserInfo struct {
	UserId      int64  `json:"userId`
	NickName    string `json:"nickname"`
	UserAccount string `json:"userAccount"`
	DomainName  string `json:"domainName"`
	UsedSize    int64  `json:"usedSize"`
	Quota       int64  `json:"quota"`
}

func (u *UserInfo) GetName() string {
	if u.DomainName != "" {
		return u.DomainName
	}
	return u.UserAccount
}
