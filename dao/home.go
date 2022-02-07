package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"189Cloud-Downloader/model"
	"189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
)

const (
	_getHomeDirListAPI = "https://cloud.189.cn/api/open/file/listFiles.action?"
)

func (d *dao) GetHomeDirList(ctx context.Context, pn, ps int, order string, folderId ...string) (dirs []*model.Dir, err error) {
	if d.token.WebLoginToken == "" {
		err = errors.New("当前还没有用户登陆！")
		return
	}
	if folderId == nil || folderId[0] == "" || folderId[0] == "~" {
		folderId = []string{"-11"}
	}
	var params = url.Values{}
	params.Set("orderBy", order)
	params.Set("folderId", folderId[0])
	params.Set("pageNum", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	params.Set("noCache", utils.GenNoCacheNum())
	var req *http.Request
	if req, err = http.NewRequest("GET", _getHomeDirListAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", _getHomeDirListAPI+params.Encode(), err)
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("COOKIE_LOGIN_USER=%s", d.token.WebLoginToken))
	req.Header.Set("accept", "application/json;charset=UTF-8")
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do(GET %s) error(%v)", _getHomeDirListAPI+params.Encode(), err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll(resp.Body) error(%v)", err)
		return
	}
	var res struct {
		ResCode    int    `json:"res_code"`
		ResMsg     string `json:"res_message"`
		FileListAO *struct {
			FileList   []*model.Dir `json:"fileList"`
			FolderList []*model.Dir `json:"folderList"`
		} `json:"fileListAO"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err = decoder.Decode(&res); err != nil {
		log.Error("decoder.Decode() error(%v)", err)
		return
	}
	if res.ResCode != 0 {
		log.Error("d.GetHomeDirList() error(%s)", res.ResMsg)
		return
	}
	for _, dir := range res.FileListAO.FolderList {
		dir.IsFolder = true
		dir.IsHome = true
		dirs = append(dirs, dir)
	}
	for _, dir := range res.FileListAO.FileList {
		dir.IsHome = true
		dirs = append(dirs, dir)
	}
	return
}

func (d *dao) GetHomeDirAll(ctx context.Context, fileID ...string) (dirs []*model.Dir, err error) {
	for pn := 1; ; pn++ {
		var dirs2 []*model.Dir
		if dirs2, err = d.GetHomeDirList(ctx, pn, 100, "filename", fileID...); err != nil {
			log.Error("d.GetHomeDirList() pn(%d) ps(100) order(filename) fileID(%+v)", pn, fileID, err)
			return
		}
		dirs = append(dirs, dirs2...)
		if len(dirs2) < 100 {
			return
		}
	}
}
