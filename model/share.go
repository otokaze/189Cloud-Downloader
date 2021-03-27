package model

type ShareInfo struct {
	ShortCode  string
	AccessCode string
	ShareID    string
	VerifyCode string
	Name       string
	IsFile     bool
}

func (share *ShareInfo) GetShortName() string {
	var name = []rune(share.Name)
	if len(name) <= 6 {
		return share.Name
	}
	return string(name[:6]) + "..."
}
