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
}
