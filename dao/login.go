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
	"time"

	"github.com/otokaze/189Cloud-Downloader/model"
	"github.com/otokaze/189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
)

type token struct {
	SessionKey          string
	SessionSecret       string
	FamilySessionKey    string
	FamilySessionSecret string
	AccessToken         string
	RefreshToken        string
	// 有效期的token
	SskAccessToken string
	// token 过期时间点，时间戳ms
	SskAccessTokenExp int64
	RsaPublicKey      string
	WebLoginToken     string
}

type appSessionResp struct {
	ResCode             int    `json:"res_code"`
	ResMessage          string `json:"res_message"`
	AccessToken         string `json:"accessToken"`
	FamilySessionKey    string `json:"familySessionKey"`
	FamilySessionSecret string `json:"familySessionSecret"`
	GetFileDiffSpan     int    `json:"getFileDiffSpan"`
	GetUserInfoSpan     int    `json:"getUserInfoSpan"`
	IsSaveName          string `json:"isSaveName"`
	KeepAlive           int    `json:"keepAlive"`
	LoginName           string `json:"loginName"`
	RefreshToken        string `json:"refreshToken"`
	SessionKey          string `json:"sessionKey"`
	SessionSecret       string `json:"sessionSecret"`
}

type appLoginParams struct {
	CaptchaToken string
	Lt           string
	ReturnUrl    string
	ParamId      string
	ReqId        string
	jRsaKey      string
}

const (
	_refSessionAPI          = "https://cloud.189.cn/ssoLogin.action?"
	_getParamsAPI           = "https://cloud.189.cn/unifyLoginForPC.action?"
	_getSessionAPI          = "https://api.cloud.189.cn/getSessionForPC.action?"
	_getLoginedInfoAPI      = "https://cloud.189.cn/v2/getLoginedInfos.action?"
	_getSsKeyAccessTokenAPI = "https://api.cloud.189.cn/open/oauth2/getAccessTokenBySsKey.action?"

	_appLoginAPI  = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	_rsakeyFormat = "-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----"
)

func (d *dao) LoginWithCookie(ctx context.Context, cookie string) (err error) {
	d.token = &token{WebLoginToken: cookie}
	return
}

func (d *dao) Login(ctx context.Context, username, password string) (user *model.UserInfo, err error) {
	var loginParams *appLoginParams
	if loginParams, err = d.getLoginParams(ctx); err != nil {
		return
	}
	d.token.RsaPublicKey = fmt.Sprintf(_rsakeyFormat, loginParams.jRsaKey)
	var toUrl string
	if toUrl, err = d.loginSubmit(ctx, loginParams, username, password); err != nil {
		return
	}
	var sess *appSessionResp
	if sess, err = d.getSessionForPC(ctx, toUrl); err != nil {
		return
	}
	d.token.SessionKey = sess.SessionKey
	d.token.SessionSecret = sess.SessionSecret
	d.token.FamilySessionKey = sess.FamilySessionKey
	d.token.FamilySessionSecret = sess.FamilySessionSecret
	d.token.AccessToken = sess.AccessToken
	d.token.RefreshToken = sess.RefreshToken
	if d.token.SskAccessToken, d.token.SskAccessTokenExp, err = d.getSsKeyAccessToken(ctx); err != nil {
		log.Error("d.getSsKeyAccessToken() sessionKey(%s) error(%v)", d.token.SessionKey, err)
		return
	}
	if d.token.WebLoginToken, err = d.RefreshCookieToken(d.token.SessionKey); err != nil {
		log.Error("d.RefreshCookieToken(%s) error(%v)", d.token.SessionKey, err)
		return
	}
	var info interface{}
	if info, err = d.GetLoginedInfo(ctx, false); err != nil {
		return
	}
	user = info.(*model.UserInfo)
	return
}

func (d *dao) getLoginParams(ctx context.Context) (p *appLoginParams, err error) {
	params := url.Values{}
	params.Set("appId", "8025431004")
	params.Set("clientType", "10020")
	params.Set("noCache", utils.GenNoCacheNum())
	params.Set("timeStamp", strconv.Itoa(int(time.Now().UTC().UnixNano()/1e6)))
	params.Set("returnURL", "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html")
	var resp *http.Response
	if resp, err = d.httpCli.Get(_getParamsAPI + params.Encode()); err != nil {
		log.Error("d.httpCli.Get(%s) error(%v)", _getParamsAPI+params.Encode(), err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll(resp.Body) error(%v)", err)
		return
	}
	var content = string(body)
	re, _ := regexp.Compile("captchaToken' value='(.+?)'")
	p = &appLoginParams{}
	p.CaptchaToken = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("lt = \"(.+?)\"")
	p.Lt = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("returnUrl = '(.+?)'")
	p.ReturnUrl = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("paramId = \"(.+?)\"")
	p.ParamId = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("reqId = \"(.+?)\"")
	p.ReqId = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("j_rsaKey\" value=\"(.+?)\"")
	p.jRsaKey = re.FindStringSubmatch(content)[1]
	return
}

func (d *dao) getSessionForPC(ctx context.Context, toUrl string) (sess *appSessionResp, err error) {
	var params = url.Values{}
	params.Set("clientType", "TELEMAC")
	params.Set("version", "1.0.0")
	params.Set("channelId", "web_cloud.189.cn")
	params.Set("redirectURL", url.QueryEscape(toUrl))
	params.Set("noCache", utils.GenNoCacheNum())
	var req *http.Request
	if req, err = http.NewRequest("GET", _getSessionAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET, %s) error(%v)", _getSessionAPI+params.Encode(), err)
		return
	}
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Get(%s) error(%v)", _getSessionAPI+params.Encode(), err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		return
	}
	sess = &appSessionResp{}
	if err = json.Unmarshal(body, sess); err != nil {
		log.Error("json.Unmarshal error(%v)", err)
		return
	}
	if sess.ResCode != 0 {
		err = errors.New("获取session失败")
		return
	}
	return
}

func (d *dao) loginSubmit(ctx context.Context, loginParams *appLoginParams, username, password string) (toUrl string, err error) {
	rsaUserName, _ := utils.RsaEncrypt([]byte(d.token.RsaPublicKey), []byte(username))
	rsaPassword, _ := utils.RsaEncrypt([]byte(d.token.RsaPublicKey), []byte(password))
	var params = url.Values{}
	params.Set("isOauth2", "false")
	params.Set("cb_SaveName", "1")
	params.Set("accountType", "02")
	params.Set("mailSuffix", "@189.cn")
	params.Set("dynamicCheck", "FALSE")
	params.Set("appKey", "8025431004")
	params.Set("clientType", "10020")
	params.Set("paramId", loginParams.ParamId)
	params.Set("returnUrl", loginParams.ReturnUrl)
	params.Set("captchaToken", loginParams.CaptchaToken)
	params.Set("userName", "{RSA}"+utils.Base64toHex(string(utils.Base64Encode(rsaUserName))))
	params.Set("password", "{RSA}"+utils.Base64toHex(string(utils.Base64Encode(rsaPassword))))
	var req *http.Request
	if req, err = http.NewRequest("POST", _appLoginAPI, strings.NewReader(params.Encode())); err != nil {
		log.Error("http.NewRequest(POST, %s) params(%s) error(%v)", _appLoginAPI, params.Encode(), err)
		return
	}
	req.Header.Set("lt", loginParams.Lt)
	req.Header.Set("REQID", loginParams.ReqId)
	req.Header.Set("Cookie", "LT="+loginParams.Lt)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://open.e.189.cn/api/logbox/oauth2/unifyAccountLogin.do")
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do error(%v)", err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll error(%v)", err)
		return
	}
	var loginResult struct {
		Result int    `json:"result"`
		Msg    string `json:"msg"`
		ToUrl  string `json:"toUrl"`
	}
	if err = json.Unmarshal(body, &loginResult); err != nil {
		log.Error("json.Unmarshal() error(%v)", err)
		return
	}
	if loginResult.Result != 0 || loginResult.ToUrl == "" {
		err = fmt.Errorf("登陆失败！code(%d) toUrl(%s)", loginResult.Result, loginResult.ToUrl)
		return
	}
	toUrl = loginResult.ToUrl
	return
}

func (d *dao) getSsKeyAccessToken(ctx context.Context) (access string, expr int64, err error) {
	var params = url.Values{}
	params.Set("sessionKey", d.token.SessionKey)
	timestamp := utils.Timestamp()
	var req *http.Request
	if req, err = http.NewRequest("GET", _getSsKeyAccessTokenAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET, %s) error(%v)", _getSsKeyAccessTokenAPI+params.Encode(), err)
		return
	}
	var signParams = map[string]string{
		"Timestamp":  strconv.Itoa(timestamp),
		"sessionKey": d.token.SessionKey,
		"AppKey":     "601102120",
	}
	req.Header.Set("Sign-Type", "1")
	req.Header.Set("AppKey", "601102120")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Timestamp", strconv.Itoa(timestamp))
	req.Header.Set("Signature", utils.SignatureOfMd5(signParams))
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do error(%v)", err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll error(%v)", err)
		return
	}
	var res struct {
		// token过期时间，默认30天
		ExpiresIn   int64  `json:"expiresIn"`
		AccessToken string `json:"accessToken"`
	}
	if err = json.Unmarshal(body, &res); err != nil {
		log.Error("json.Unmarshal() error(%v)", err)
		return
	}
	expr = res.ExpiresIn
	access = res.AccessToken
	return
}

func (d *dao) RefreshCookieToken(sessKey string) (loginCookie string, err error) {
	var params = url.Values{}
	params.Set("sessionKey", sessKey)
	params.Set("redirectUrl", "main.action")
	var httpCli = *d.httpCli
	httpCli.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	var resp *http.Response
	if resp, err = httpCli.Get(_refSessionAPI + params.Encode()); err != nil {
		log.Error("httpCli.Get(%s) error(%v)", _refSessionAPI+params.Encode(), err)
		return
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "COOKIE_LOGIN_USER" {
			loginCookie = cookie.Value
			break
		}
	}
	if loginCookie == "" {
		err = errors.New("Cookie is empty")
		return
	}
	return
}

func (d *dao) GetLoginedInfo(ctx context.Context, returnJSON ...bool) (info interface{}, err error) {
	if d.token.WebLoginToken == "" {
		err = errors.New("当前还没有登陆！")
		return
	}
	var params = url.Values{}
	params.Set("showPC", "true")
	params.Set("noCache", utils.GenNoCacheNum())
	var req *http.Request
	if req, err = http.NewRequest("GET", _getLoginedInfoAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET %s) error(%v)", _getLoginedInfoAPI+params.Encode(), err)
		return
	}
	req.Header.Set("Cookie", fmt.Sprintf("COOKIE_LOGIN_USER=%s", d.token.WebLoginToken))
	var resp *http.Response
	if resp, err = d.httpCli.Do(req); err != nil {
		log.Error("d.httpCli.Do() error(%v)", err)
		return
	}
	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll(resp.Body) error(%v)", err)
		return
	}
	if len(returnJSON) > 0 && returnJSON[0] {
		info = string(body)
		return
	}
	var userinfo = &model.UserInfo{}
	if err = json.Unmarshal(body, userinfo); err != nil {
		log.Error("json.Unmarshal(body) error(%v)", err)
		return
	}
	info = userinfo
	return
}

func (d *dao) Logout(ctx context.Context) (err error) {
	d.token = &token{}
	return
}
