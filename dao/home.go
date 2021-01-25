package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/otokaze/189Cloud-Downloader/model"
	"github.com/otokaze/189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
)

const (
	_getHomeDirListAPI = "https://cloud.189.cn/v2/listFiles.action?"
)

func (d *dao) GetHomeDirList(ctx context.Context, pn, ps int, order string, fileID ...string) (dirs []*model.Dir, paths []*model.Path, err error) {
	if d.token.WebLoginToken == "" {
		err = errors.New("当前还没有用户登陆！")
		return
	}
	if fileID == nil || fileID[0] == "" || fileID[0] == "~" {
		fileID = []string{"-11"}
	}
	var params = url.Values{}
	params.Set("order", order)
	params.Set("fileId", fileID[0])
	params.Set("pageNum", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	params.Set("noCache", utils.GenNoCacheNum())
	var req *http.Request
	if req, err = http.NewRequest("GET", _getHomeDirListAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", _getHomeDirListAPI+params.Encode(), err)
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("COOKIE_LOGIN_USER=%s", d.token.WebLoginToken))
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

func (d *dao) GetHomeDirAll(ctx context.Context, fileID ...string) (dirs []*model.Dir, paths []*model.Path, err error) {
	for pn := 1; ; pn++ {
		var dirs2 []*model.Dir
		if dirs2, paths, err = d.GetHomeDirList(ctx, pn, 100, "ASC", fileID...); err != nil {
			log.Error("d.GetHomeDirList() pn(%d) ps(100) order(ASC) fileID(%+v)", pn, fileID, err)
			return
		}
		dirs = append(dirs, dirs2...)
		if len(dirs2) <= 100 {
			return
		}
	}
}
