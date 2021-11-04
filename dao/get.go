package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	URL "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"189Cloud-Downloader/model"
	"189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
	"github.com/otokaze/go-kit/progressbar"
)

const (
	_getDownloadUrlAPI = "https://cloud.189.cn/api/open/file/getFileDownloadUrl.action?"
)

func (d *dao) GetDownloadURL(ctx context.Context, fileId string, share ...*model.ShareInfo) (URL string, err error) {
	var params = url.Values{}
	params.Set("dt", "1")
	params.Set("fileId", fileId)
	params.Set("noCache", utils.GenNoCacheNum())
	if len(share) > 0 && share[0] != nil {
		params.Set("shareId", strconv.FormatInt(share[0].ShareID, 10))
	}
	var req *http.Request
	if req, err = http.NewRequest("GET", _getDownloadUrlAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", _getDownloadUrlAPI, err)
		return
	}
	req.Header.Set("accept", "application/json;charset=UTF-8")
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
	var res struct {
		ResCode         int    `json:"res_code"`
		ResMsg          string `json:"res_message"`
		FileDownloadUrl string `json:"fileDownloadUrl"`
	}
	if err = json.Unmarshal(body, &res); err != nil {
		log.Error("json.Unmarshal() error(%v)", err)
		return
	}
	if res.ResCode != 0 {
		log.Error("d.GetDownloadURL() error(%s)", res.ResMsg)
		return
	}
	URL = res.FileDownloadUrl
	return
}

func (d *dao) Download(ctx context.Context, url, toPath string, c int, tmpDirs ...string) (err error) {
	if err = os.MkdirAll(toPath, 0777); err != nil {
		log.Error("os.MkdirAll(%s, 0777) error(%v)", toPath, err)
		return
	}
	var tmpPath string
	if len(tmpDirs) == 0 || tmpDirs[0] == "" {
		tmpPath = os.TempDir()
	} else {
		tmpPath = strings.TrimRight(tmpDirs[0], "/")
	}
	tmpPath = tmpPath + "/.downloading"
	if err = os.MkdirAll(tmpPath, 0777); err != nil {
		log.Error("os.MkdirAll(%s, 0777) error(%v)", tmpPath, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", url, err)
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("COOKIE_LOGIN_USER=%s", d.token.WebLoginToken))
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do(req) error(%v)", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("resp.StatusCode(%d) is not OK(200)", resp.StatusCode)
		return
	}
	var b int64
	if b, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err != nil {
		log.Error("strconv.ParseInt(%s) error(%v)", resp.Header.Get("Content-Length"), err)
		return
	}
	var disp string
	if disp, err = URL.QueryUnescape(resp.Header.Get("Content-Disposition")); err != nil {
		log.Error("URL.QueryUnescape(%s) error(%v)", resp.Header.Get("Content-Disposition"), err)
		return
	}
	var (
		matchs   []string
		fnameReg = regexp.MustCompile(`attachment;filename="(.*?)"`)
	)
	if matchs = fnameReg.FindStringSubmatch(disp); len(matchs) <= 1 {
		err = fmt.Errorf("Content-Disposition: %s, 没有找到文件名！", disp)
		return
	}
	var shortName string
	if r := []rune(matchs[1]); len(r) <= 12 {
		shortName = matchs[1]
	} else {
		shortName = string(r[:12]) + "..."
	}
	if c != 1 && b < 10*1024*1024 {
		c = 1
	}
	var tmpDir string
	if tmpDir, err = ioutil.TempDir(tmpPath, matchs[1]); err != nil {
		log.Error("ioutil.TempDir(%s, %s)", tmpPath, matchs[1])
		return
	}
	var bar = progressbar.New(nil)
	bar.SetMax(b)
	bar.SetPrefix(shortName)
	bar.SetSuffix("下载中...")
	bar.Run()
	var wg sync.WaitGroup
	for i := 0; i < c; i++ {
		wg.Add(1)
		start := b / int64(c) * int64(i)
		end := b / int64(c) * int64(i+1)
		if i == c-1 {
			end = b
		}
		if start > 0 {
			start++
		}
		go func(i int, start, end int64) {
			defer wg.Done()
			var (
				retry = -1
				size  int64
			)
		download:
			if retry++; retry >= 3 {
				println()
				err = errors.New("下载失败！")
				log.Error("file(%s) part(%d) 下载失败！可能当前资源并不支持多线程下载！", matchs[1], i)
				bar.Stop()
				return
			} else if retry > 0 {
				println()
				log.Error("file(%s) part(%d) 下载失败！正在进行重试...（%d/3）", matchs[1], i, retry)
				bar.Add(-size)
				size = 0
				time.Sleep(3 * time.Second)
			}
			var downReq *http.Request
			if downReq, err = http.NewRequest("GET", resp.Request.URL.String(), nil); err != nil {
				log.Error("http.NewRequest(GET %s) error(%v)", resp.Request.URL.String(), err)
				goto download
			}
			downReq.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			var downResp *http.Response
			if downResp, err = d.httpCli.Do(downReq); err != nil {
				log.Error("d.httpCli.Do(downReq) error(%v)", err)
				goto download
			}
			var tmpFile *os.File
			if tmpFile, err = os.Create(fmt.Sprintf("%s/%s.%d", tmpDir, matchs[1], i)); err != nil {
				log.Error("os.Create(%s/%s.%d) error(%v)", tmpDir, matchs[1], i, err)
				downResp.Body.Close()
				goto download
			}
			time.Sleep(time.Duration(i) * time.Second)
			var written int64
			if written, err = d.readTo(tmpFile, downResp.Body, bar); err != nil || written < end-start {
				if bar.IsRunning() {
					downResp.Body.Close()
					tmpFile.Close()
					goto download
				}
				return
			}
			return
		}(i, start, end)
	}
	wg.Wait()
	if !bar.IsRunning() {
		return
	}
	var target *os.File
	if target, err = os.Create(toPath + "/" + matchs[1]); err != nil {
		log.Error("os.Create(%s/%s) error(%v)", toPath, matchs[1], err)
		return
	}
	defer target.Close()
	bar.SetSuffix("合并中...")
	bar.Set(0)
	for i := 0; i < c; i++ {
		var part *os.File
		if part, err = os.Open(fmt.Sprintf("%s/%s.%d", tmpDir, matchs[1], i)); err != nil {
			log.Error("os.Open(%s/%s.%d) 读取下载文件分片时出错：%v", tmpDir, matchs[1], i, err)
			return
		}
		if _, err = d.readTo(target, part, bar); err != nil {
			log.Error("d.readTo(target, part) error(%v)", err)
			return
		}
	}
	bar.Set(b)
	bar.SetSuffix("下载完成")
	os.RemoveAll(tmpDir)
	bar.Stop()
	return
}

func (d *dao) readTo(dst io.Writer, src io.Reader, bar ...*progressbar.Bar) (written int64, err error) {
	var buf = make([]byte, 4096)
	for {
		if len(bar) > 0 && !bar[0].IsRunning() {
			return
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			var w int
			if w, err = dst.Write(buf[0:n]); err != nil {
				log.Error("dst.Write(buf[0:%d]) error(%v)", n, err)
				break
			}
			if n != w {
				err = io.ErrShortWrite
				log.Error("dst.Write(buf[0:%d]) error(%v)", n, io.ErrShortWrite)
				break
			}
			if w > 0 {
				written += int64(w)
				if len(bar) > 0 {
					bar[0].Add(int64(w))
				}
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				log.Error("d.readTo(target, part) error(%v)", readErr)
				err = readErr
			}
			break
		}
	}
	return
}
