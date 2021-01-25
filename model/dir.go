package model

type Path struct {
	FileId   string `json:"fileId"`
	FileName string `json:"fileName"`
}

func (p *Path) GetShortName() string {
	var name = []rune(p.FileName)
	if len(name) <= 6 {
		return p.FileName
	}
	return string(name[:6]) + "..."
}

type Dir struct {
	DownloadUrl  string `json:"downloadUrl"`
	FileIdDigest string `json:"fileIdDigest"`
	CreateTime   string `json:"createTime"`
	FileID       string `json:"fileId"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	FileType     string `json:"fileType"`
	IsFolder     bool   `json:"isFolder"`
	LastOpTime   string `json:"lastOpTime"`
	ParentID     string `json:"parentId"`
}

type PathTree []*Path

func (p PathTree) GetCurrentPath() *Path {
	if len(p) > 0 {
		return p[len(p)-1]
	}
	return nil
}

func (p PathTree) GetParentPath() *Path {
	if len(p) > 1 {
		return p[len(p)-2]
	}
	return nil
}

func (p PathTree) GetRootPath() *Path {
	if len(p) > 0 {
		return p[0]
	}
	return nil
}

type Dirs []*Dir

func (ds Dirs) Find(fileId string) *Dir {
	for _, d := range ds {
		if d.FileID == fileId {
			return d
		}
	}
	return nil
}

func (d *Dir) IsPrivate() bool {
	if d.FileIdDigest != "" {
		return true
	}
	return false
}

func (d *Dir) IsPublic() bool {
	if d.FileIdDigest != "" {
		return false
	}
	return true
}
