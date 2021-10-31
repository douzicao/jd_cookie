package jd_cookie

import (
	"fmt"
	"strings"
	"time"

	"github.com/douzicao/sillyGirl/core"
	"github.com/douzicao/sillyGirl/develop/qinglong"
	"github.com/gin-gonic/gin"
)

var pinWX = core.NewBucket("pinWX")
var pinTG = core.NewBucket("pinTG")
var pinWXMP = core.NewBucket("pinWXMP")


func init() {
	//
	core.Server.POST("/cookie", func(c *gin.Context) {
		cookie := c.Query("ck")
		ck := &JdCookie{
			PtKey: core.FetchCookieValue(cookie, "pt_key"),
			PtPin: core.FetchCookieValue(cookie, "pt_pin"),
		}
		type Result struct {
			Code    int         `json:"code"`
			Data    interface{} `json:"data"`
			Message string      `json:"message"`
		}
		result := Result{
			Data: nil,
			Code: 300,
		}
		if ck.PtPin == "" || ck.PtKey == "" {
			result.Message = "一句mmp，不知当讲不当讲。"
			c.JSON(200, result)
			return
		}
		if !ck.Available() {
			result.Message = "无效的ck，请重试。"
			c.JSON(200, result)
			return
		}
		value := fmt.Sprintf(`pt_key=%s;\s*?pt_pin=%s;`, ck.PtKey, ck.PtPin)
		envs, err := qinglong.GetEnvs("JD_COOKIE")
		if err != nil {
			result.Message = err.Error()
			c.JSON(200, result)
			return
		}
		find := false
		for _, env := range envs {
			if strings.Contains(env.Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) {
				envs = []qinglong.Env{env}
				find = true
				break
			}
		}
		if !find {

			if err := qinglong.AddEnv(qinglong.Env{
				Name:  "JD_COOKIE",
				Value: value,
			}); err != nil {
				result.Message = err.Error()
				c.JSON(200, result)
				return
			}
			rt := ck.Nickname + "，添加成功。"
			core.NotifyMasters(rt)
			result.Message = rt
			result.Code = 200
			c.JSON(200, result)
			return
		} else {
			env := envs[0]
			env.Value = value
			if env.Status != 0 {
				if err := qinglong.Config.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
					result.Message = err.Error()
					c.JSON(200, result)
					return
				}
				env.Status = 0
				if err := qinglong.UdpEnv(env); err != nil {
					result.Message = err.Error()
					c.JSON(200, result)
					return
				}
			}
			rt := ck.Nickname + "，更新成功。"
			core.NotifyMasters(rt)
			result.Message = rt
			result.Code = 200
			c.JSON(200, result)
			return
		}
	})
	core.AddCommand("jd", []core.Function{
		{
			Rules: []string{`unbind ?`},
			Handle: func(s core.Sender) interface{} {
				s.Disappear(time.Second * 40)
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					return "暂时无法操作。"
				}
				for _, env := range envs {
					pt_pin := FetchJdCookieValue("pt_pin", env.Value)
					pin(s.GetImType()).Foreach(func(k, v []byte) error {
						if string(k) == pt_pin && string(v) == s.Get() {
							s.Reply(fmt.Sprintf("已解绑，%s。", pt_pin))
							defer func() {
								pinWX.Set(string(k), "")
							}()
						}
						return nil
					})
				}
				return "操作完成"
			},
		},
		{
			Rules:   []string{`raw pt_key=([^;=\s]+);\s*pt_pin=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s core.Sender) interface{} {
				s.Reply(s.Delete())
				s.Disappear(time.Second * 20)
				for _, v := range s.GetAllMatch() {
					ck := &JdCookie{
						PtKey: v[0],
						PtPin: v[1],
					}
					if len(ck.PtKey) <= 20 {
						s.Reply("再捣乱我就报警啦！") //
						continue
					}
					if !ck.Available() {
						s.Reply("无效的账号。") //有瞎编ck的嫌疑
						continue
					}
					if ck.Nickname == "" {
						s.Reply("请修改昵称！")
					}

					value := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
					envs, err := qinglong.GetEnvs("JD_COOKIE")
					if err != nil {
						s.Reply(err)
						continue
					}
					find := false
					for _, env := range envs {
						if strings.Contains(env.Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) {
							envs = []qinglong.Env{env}
							find = true
							break
						}
					}
					pin(s.GetImType()).Set(ck.PtPin, s.GetUserID())
					if !find {
						if err := qinglong.AddEnv(qinglong.Env{
							Name:  "JD_COOKIE",
							Value: value,
						}); err != nil {
							s.Reply(err)
							continue
						}
						rt := ck.Nickname + "，添加成功。"
						core.NotifyMasters(rt)
						s.Reply(rt)
						continue
					} else {
						env := envs[0]
						env.Value = value
						if env.Status != 0 {
							if err := qinglong.Config.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
								s.Reply(err)
								continue
							}
						}
						env.Status = 0
						if err := qinglong.UdpEnv(env); err != nil {
							s.Reply(err)
							continue
						}
						rt := ck.Nickname + "，更新成功。"
						core.NotifyMasters(rt)
						s.Reply(rt)
						continue
					}
				}
				return nil
		},
    	{
			Rules:   []string{`raw pin=([^;=\s]+);\s*wskey=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s core.Sender) interface{} {
				s.Reply(s.Delete())
				s.Disappear(time.Second * 20)
				value := fmt.Sprintf("pin=%s;wskey=%s;", s.Get(0), s.Get(1))

				pt_key, err := getKey(value)
				if err == nil {
					if strings.Contains(pt_key, "fake") {
						return "无效的wskey，请重试。"
					}
				} else {
					s.Reply(err)
				}
				ck := &JdCookie{
					PtKey: pt_key,
					PtPin: s.Get(0),
				}
				ck.Available()
				envs, err := qinglong.GetEnvs("pin=")
				if err != nil {
					return err
				}
				pin(s.GetImType()).Set(ck.PtPin, s.GetUserID())
				var envCK *qinglong.Env
				var envWsCK *qinglong.Env
				for i := range envs {
					if strings.Contains(envs[i].Value, fmt.Sprintf("pin=%s;wskey=", ck.PtPin)) && envs[i].Name == "JD_WSCK" {
						envWsCK = &envs[i]
					} else if strings.Contains(envs[i].Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) && envs[i].Name == "JD_COOKIE" {
						envCK = &envs[i]
					}
				}
				value2 := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
				if envCK == nil {
					qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value2,
					})
				} else {
					if envCK.Status != 0 {
						envCK.Value = value2
						if err := qinglong.UdpEnv(*envCK); err != nil {
							return err
						}
					}
				}
				if envWsCK == nil {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_WSCK",
						Value: value,
					}); err != nil {
						return err
					}
					return ck.Nickname + ",添加成功。"
				} else {
					envWsCK.Value = value
					if envWsCK.Status != 0 {
						if err := qinglong.Config.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+envWsCK.ID+`"]`)); err != nil {
							return err
						}
					}
					envWsCK.Status = 0
					if err := qinglong.UdpEnv(*envWsCK); err != nil {
						return err
					}
					return ck.Nickname + ",更新成功。"
				}
			},
		},
	})
}


type AutoGenerated struct {
	ClientVersion string `json:"clientVersion"`
	Client        string `json:"client"`
	Sv            string `json:"sv"`
	St            string `json:"st"`
	UUID          string `json:"uuid"`
	Sign          string `json:"sign"`
	FunctionID    string `json:"functionId"`
}

func getSign() *AutoGenerated {
	var a *AutoGenerated
	jdWSCK.Foreach(func(_, v []byte) error {
		t := &AutoGenerated{}
		if json.Unmarshal(v, t) == nil {
			if a == nil || t.St > a.St {
				a = t
			}
		}
		return nil
	})
	if a != nil {
		a.FunctionID = "genToken"
	}
	return a
}

func getKey(WSCK string) (string, error) {
	v := url.Values{}
	s := getSign()
	v.Add("functionId", s.FunctionID)
	v.Add("clientVersion", s.ClientVersion)
	v.Add("client", s.Client)
	v.Add("uuid", s.UUID)
	v.Add("st", s.St)
	v.Add("sign", s.Sign)
	v.Add("sv", s.Sv)
	req := httplib.Post(`https://api.m.jd.com/client.action?` + v.Encode())
	req.Header("cookie", WSCK)
	req.Header("User-Agent", ua2)
	req.Header("content-type", `application/x-www-form-urlencoded; charset=UTF-8`)
	req.Header("charset", `UTF-8`)
	req.Header("accept-encoding", `br,gzip,deflate`)
	req.Body(`body=%7B%22action%22%3A%22to%22%2C%22to%22%3A%22https%253A%252F%252Fplogin.m.jd.com%252Fcgi-bin%252Fm%252Fthirdapp_auth_page%253Ftoken%253DAAEAIEijIw6wxF2s3bNKF0bmGsI8xfw6hkQT6Ui2QVP7z1Xg%2526client_type%253Dandroid%2526appid%253D879%2526appup_type%253D1%22%7D&`)
	data, err := req.Bytes()
	if err != nil {
		return "", err
	}
	tokenKey, _ := jsonparser.GetString(data, "tokenKey")
	pt_key, err := appjmp(tokenKey)
	if err != nil {
		return "", err
	}
	return pt_key, nil
}

func appjmp(tokenKey string) (string, error) {
	v := url.Values{}
	v.Add("tokenKey", tokenKey)
	v.Add("to", ``)
	v.Add("client_type", "android")
	v.Add("appid", "879")
	v.Add("appup_type", "1")
	req := httplib.Get(`https://un.m.jd.com/cgi-bin/app/appjmp?` + v.Encode())
	req.Header("User-Agent", ua2)
	req.Header("accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3`)
	req.SetCheckRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	rsp, err := req.Response()
	if err != nil {
		return "", err
	}
	cookies := strings.Join(rsp.Header.Values("Set-Cookie"), " ")
	pt_key := core.FetchCookieValue(cookies, "pt_key")
	return pt_key, nil
}
