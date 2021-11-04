package model

import (
	"fmt"
)

type Dir struct {
	CreateDate string `json:"createDate"`
	FileCata   int    `json:"fileCata"`
	// ψ(｀∇´)ψ 无法吐槽天翼云的工程师，在相同结构下这id字段既可以是number又可以是string。666
	ID interface{} `json:"id"`
	// ψ(｀∇´)ψ 同上！
	ParentID     interface{} `json:"parentId"`
	LastOpTime   string      `json:"lastOpTime"`
	MediaType    int         `json:"mediaType"`
	Md5          string      `json:"md5"`
	Name         string      `json:"name"`
	Rev          string      `json:"rev"`
	Size         int64       `json:"size"`
	StarLabel    int         `json:"starLabel"`
	FileListSize int64       `json:"fileListSize"`
	IsFolder     bool
	IsHome       bool
}

func (dir *Dir) GetShortName() string {
	var name = []rune(dir.Name)
	if len(name) <= 12 {
		return dir.Name
	}
	return string(name[:12]) + "..."
}

func (dir *Dir) GetID() string {
	return fmt.Sprintf("%s", dir.ID)
}

func (dir *Dir) GetParentID() string {
	return fmt.Sprintf("%s", dir.ParentID)
}
