package jd_cookie

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/douzicao/sillyGirl/core"
	"github.com/douzicao/sillyGirl/develop/qinglong"
	"golang.org/x/net/proxy"
)

func init() {
	if !core.Bucket("qinglong").GetBool("enable_qinglong", true) {
		return
	}
	data, _ := os.ReadFile("dev.go")
	if !strings.Contains(string(data), "jd_cookie") && !jd_cookie.GetBool("enable_jd_cookie") {
		return
	}
	initAsset()
	initCheck()
	initEnv()
	initLogin()
	initSubmit()
	initNotify()
	buildHttpTransportWithProxy()
	if Transport != nil {
		logs.Info("douzicao")
	} else {
		logs.Info("douzicao")
	}
	logs.Info(
		"douzicao%s",
		`douzicao`,
	)
}

var Transport *http.Transport

func buildHttpTransportWithProxy() {
	addr := jd_cookie.Get("http_proxy")
	if strings.Contains(addr, "http://") {
		if addr != "" {
			u, err := url.Parse(addr)
			if err != nil {
				logs.Warn("can't connect to the http proxy:", err)
				return
			}
			Transport = &http.Transport{Proxy: http.ProxyURL(u)}
		}
	}
	if strings.Contains(addr, "sock5://") || strings.Contains(addr, "socks5://") {
		addr = strings.Replace(addr, "sock5://", "", -1)
		addr = strings.Replace(addr, "socks5://", "", -1)
		var auth *proxy.Auth
		v := strings.Split(addr, "@")
		if len(v) == 3 {
			auth = &proxy.Auth{
				User:     v[1],
				Password: v[2],
			}
			addr = v[0]
		}
		dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
		if err != nil {
			logs.Warn("can't connect to the sock5 proxy:", err)
			return
		}
		Transport = &http.Transport{
			Dial: dialer.Dial,
		}
	}
}

func GetEnvs(ql *qinglong.QingLong, s string) ([]qinglong.Env, error) {
	envs, err := qinglong.GetEnvs(ql, s)
	if err != nil {
		if s == "JD_COOKIE" {
			i := 0
			for _, env := range envs {
				if env.Status == 0 {
					i++
				}
			}
			ql.SetNumber(i)
		}
	}
	return envs, err
}
