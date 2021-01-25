package model

type ShareInfo struct {
	ShortCode  string
	PassCode   string
	ShareID    string
	VerifyCode string
	Name       string
}

func (share *ShareInfo) GetShortName() string {
	var name = []rune(share.Name)
	if len(name) <= 6 {
		return share.Name
	}
	return string(name[:6]) + "..."
}
