package model

type ShareInfo struct {
	ShareID    int64
	AccessCode string
	ShareCode  string
	FileID     string
	FileName   string
	IsFolder   bool
}
