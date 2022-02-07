package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"189Cloud-Downloader/model"
	"189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
)

const (
	_listShareDirAPI    = "https://cloud.189.cn/api/open/share/listShareDir.action?"
	_getShareInfoAPI    = "https://cloud.189.cn/api/open/share/getShareInfoByCodeV2.action?"
	_checkAccessCodeAPI = "https://cloud.189.cn/api/open/share/checkAccessCode.action?"
)

func (d *dao) GetShareDirList(ctx context.Context, share *model.ShareInfo, pn, ps int, order string, folderId ...string) (dirs []*model.Dir, err error) {
	if share == nil {
		err = errors.New("当前还没有载入任何分享链接！")
		return
	}
	if folderId == nil {
		folderId = []string{""}
	}
	var params = url.Values{}
	params.Set("shareMode", "1")
	params.Set("orderBy", order)
	params.Set("fileId", folderId[0])
	params.Set("shareDirFileId", folderId[0])
	params.Set("isFolder", strconv.FormatBool(share.IsFolder))
	params.Set("shareId", strconv.FormatInt(share.ShareID, 10))
	params.Set("accessCode", share.AccessCode)
	params.Set("pageNum", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	var req *http.Request
	if req, err = http.NewRequest("GET", _listShareDirAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET, %s) error(%v)", _listShareDirAPI+params.Encode(), err)
		return
	}
	req.Header.Set("accept", "application/json;charset=UTF-8")
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do(%+v) 请求失败！error(%v)", req, err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll error(%v)", err)
		return
	}
	body = bytes.TrimSpace(body)
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
		dirs = append(dirs, dir)
	}
	dirs = append(dirs, res.FileListAO.FileList...)
	return
}

func (d *dao) GetShareDirAll(ctx context.Context, share *model.ShareInfo, fileID ...string) (dirs []*model.Dir, err error) {
	for pn := 1; ; pn++ {
		var dirs2 []*model.Dir
		if dirs2, err = d.GetShareDirList(ctx, share, pn, 100, "filename", fileID...); err != nil {
			log.Error("d.GetShareDirList() pn(%d) ps(100) order(filename) fileID(%+v)", pn, fileID, err)
			return
		}
		dirs = append(dirs, dirs2...)
		if len(dirs2) < 100 {
			return
		}
	}
}

func (d *dao) GetShareInfo(ctx context.Context, shareCode, accessCode string) (info *model.ShareInfo, err error) {
	var params = url.Values{}
	params.Set("shareCode", shareCode)
	params.Set("noCache", utils.GenNoCacheNum())
	var req *http.Request
	if req, err = http.NewRequest("GET", _getShareInfoAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET, %s) error(%v)", _getShareInfoAPI+params.Encode(), err)
		return
	}
	req.Header.Set("accept", "application/json;charset=UTF-8")
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("httpCli.Do(%s) 请求失败！error(%v)", _getShareInfoAPI+params.Encode(), err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll error(%v)", err)
		return
	}
	var res struct {
		ResCode        int    `json:"res_code"`
		ResMsg         string `json:"res_message"`
		FileId         string `json:"fileId"`
		ShareId        int64  `json:"shareId"`
		FileName       string `json:"fileName"`
		IsFolder       bool   `json:"isFolder"`
		NeedAccessCode int8   `json:"needAccessCode"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err = decoder.Decode(&res); err != nil {
		log.Error("decoder.Decode() error(%v)", err)
		return
	}
	if res.ResCode != 0 {
		log.Error("d.GetShareFileInfo() error(%s)", res.ResMsg)
		return
	}
	if res.NeedAccessCode == 1 {
		if accessCode == "" {
			err = errors.New("该分享链接需要访问密码，否则无法读取。")
			return
		}
		params = url.Values{}
		params.Set("shareCode", shareCode)
		params.Set("accessCode", accessCode)
		params.Set("noCache", utils.GenNoCacheNum())
		var checkReq *http.Request
		if checkReq, err = http.NewRequest("GET", _checkAccessCodeAPI+params.Encode(), nil); err != nil {
			log.Error("http.NewRequest(GET, %s)", _checkAccessCodeAPI+params.Encode(), err)
			return
		}
		checkReq.Header.Set("accept", "application/json;charset=UTF-8")
		var checkResp *http.Response
		if checkResp, err = d.httpCli.Do(checkReq); err != nil {
			log.Error("httpCli.Get(%s) 请求失败！error(%v)", _checkAccessCodeAPI+params.Encode(), err)
			return
		}
		defer checkResp.Body.Close()
		if body, err = ioutil.ReadAll(checkResp.Body); err != nil {
			log.Error("ioutil.ReadAll error(%v)", err)
			return
		}
		var checkRes struct {
			ResCode int    `json:"res_code"`
			ResMsg  string `json:"res_message"`
			ShareId int64  `json:"shareId"`
		}
		if err = json.Unmarshal(body, &checkRes); err != nil {
			log.Error("json.Unmarshal() error(%v)", err)
			return
		}
		if checkRes.ResCode != 0 {
			log.Error("req checkAccessCodeAPI error(%s)", res.ResMsg)
			return
		}
		res.ShareId = checkRes.ShareId
	}
	info = &model.ShareInfo{
		ShareCode:  shareCode,
		AccessCode: accessCode,
		FileName:   res.FileName,
		IsFolder:   res.IsFolder,
		FileID:     res.FileId,
		ShareID:    res.ShareId,
	}
	return
}
