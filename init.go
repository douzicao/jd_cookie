package jd_cookie

import (
	"os"
	"strings"
)

func init() {
	data, _ := os.ReadFile("dev.go")
	if !strings.Contains(string(data), "jd_cookie") && !jd_cookie.GetBool("enable_jd_cookie") {
		return
	}
	initAsset()
	initCheck()
	initEnEn()
	initEnv()
	initHelp()
	initLogin()
	initSubmit()
	initTyt()
}
