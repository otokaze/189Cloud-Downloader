package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/otokaze/189Cloud-Downloader/model"
	"github.com/otokaze/189Cloud-Downloader/utils"
	"github.com/otokaze/go-kit/log"
)

const (
	_listShareDirAPI   = "https://cloud.189.cn/v2/listShareDirByShareIdAndFileId.action?"
	_getDownloadUrlAPI = "https://cloud.189.cn/v2/getFileDownloadUrl.action?"
)

var (
	_shareIdReg    = regexp.MustCompile(`var\s+_shareId\s+?=\s+?'(\d+)';`)
	_verifyCodeReg = regexp.MustCompile(`var\s+_verifyCode\s+?=\s+?'(\d+)';`)
	_shortCodeReg  = regexp.MustCompile(`https://cloud.189.cn/t/((?:\w+){12})`)
	_shareNameReg  = regexp.MustCompile(`<title>\s+(.*?)\s+</title>`)
)

func (d *dao) GetShareInfo(ctx context.Context, url string) (share *model.ShareInfo, err error) {
	share = &model.ShareInfo{}
	var resp *http.Response
	if resp, err = d.httpCli.Get(url); err != nil {
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	var matchShortCode []string
	if matchShortCode = _shortCodeReg.FindStringSubmatch(url); len(matchShortCode) <= 1 {
		err = errors.New("分享链接格式不正确！正确格式为：https://cloud.189.cn/t/(\\w+){12}")
		return
	}
	share.ShortCode = matchShortCode[1]
	var matchShareID []string
	if matchShareID = _shareIdReg.FindStringSubmatch(string(body)); len(matchShareID) <= 1 {
		err = errors.New("没能找到shareId，需要作者更新脚本。。。")
		return
	}
	share.ShareID = matchShareID[1]
	var matchVerifyCode []string
	if matchVerifyCode = _verifyCodeReg.FindStringSubmatch(string(body)); len(matchVerifyCode) <= 1 {
		err = errors.New("没能找到verifyCode，需要作者更新脚本。。。")
		return
	}
	share.VerifyCode = matchVerifyCode[1]
	var matchShareName []string
	if matchShareName = _shareNameReg.FindStringSubmatch(string(body)); len(matchShareName) <= 1 {
		return
	}
	share.Name = strings.Split(matchShareName[1], " ")[0]
	return
}

func (d *dao) GetShareDirList(ctx context.Context, share *model.ShareInfo, pn, ps int, order string, fileID ...string) (dirs []*model.Dir, paths []*model.Path, err error) {
	if share == nil {
		err = errors.New("当前还没有载入任何分享链接！")
		return
	}
	if fileID == nil {
		fileID = []string{""}
	}
	var params = url.Values{
		"shortCode":  []string{share.ShortCode},
		"verifyCode": []string{share.VerifyCode},
		"pageNum":    []string{strconv.Itoa(pn)},
		"pageSize":   []string{strconv.Itoa(ps)},
		"fileId":     []string{fileID[0]},
		"order":      []string{order},
		"orderBy":    []string{"1"},
	}
	var resp *http.Response
	if resp, err = d.httpCli.Get(_listShareDirAPI + params.Encode()); err != nil {
		log.Error("httpCli.Get(%s) 请求失败！error(%v)", _listShareDirAPI, err)
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll error(%v)", err)
		return
	}
	var res struct {
		Data []*model.Dir  `json:"data"`
		Path []*model.Path `json:"path"`
	}
	if err = json.Unmarshal(body, &res); err != nil {
		log.Error("json.Unmarshal() error(%v)", err)
		return
	}
	dirs, paths = res.Data, res.Path
	return
}

func (d *dao) GetShareDirAll(ctx context.Context, share *model.ShareInfo, fileID ...string) (dirs []*model.Dir, paths []*model.Path, err error) {
	for pn := 1; ; pn++ {
		var dirs2 []*model.Dir
		if dirs2, paths, err = d.GetShareDirList(ctx, share, pn, 100, "ASC", fileID...); err != nil {
			log.Error("d.GetShareDirList() pn(%d) ps(100) order(ASC) fileID(%+v)", pn, fileID, err)
			return
		}
		dirs = append(dirs, dirs2...)
		if len(dirs2) <= 100 {
			return
		}
	}
}

func (d *dao) GetDownloadURLFromShare(ctx context.Context, share *model.ShareInfo, fileId, subFileId string) (URL string, err error) {
	var params = url.Values{}
	params.Set("shortCode", share.ShortCode)
	params.Set("fileId", fileId)
	params.Set("subFileId", subFileId)
	params.Set("noCache", utils.GenNoCacheNum())
	params.Set("accessCode", "undefined")
	var req *http.Request
	if req, err = http.NewRequest("GET", _getDownloadUrlAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", _getDownloadUrlAPI, err)
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("COOKIE_LOGIN_USER=%s", d.token.WebLoginToken))
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do(req) error(%v)", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("resp.StatusCode(%d) is not OK(200)", resp.StatusCode)
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll(resp.Body) error(%v)", err)
		return
	}
	if len(body) == 0 {
		err = errors.New("resp.Body is empty")
		return
	}
	URL = strings.ReplaceAll(string(body), "\\/", "/")
	URL = strings.Trim(URL, `"`)
	if strings.HasPrefix(URL, "//") {
		URL = "https:" + URL
	}
	return
}
