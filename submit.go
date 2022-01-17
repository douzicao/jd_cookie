package jd_cookie

import (
	"fmt"
	"strings"
	"time"

	"github.com/douzicao/sillyGirl/core"
	"github.com/douzicao/sillyGirl/develop/qinglong"
)

var pinQQ = core.NewBucket("pinQQ")
var pinTG = core.NewBucket("pinTG")
var pinWXMP = core.NewBucket("pinWXMP")
var pinWX = core.NewBucket("pinWX")
var pin = func(class string) core.Bucket {
	return core.Bucket("pin" + strings.ToUpper(class))
}

func initSubmit() {
	//
	// core.Server.POST("/cookie", func(c *gin.Context) {
	// 	cookie := c.Query("ck")
	// 	ck := &JdCookie{
	// 		PtKey: core.FetchCookieValue(cookie, "pt_key"),
	// 		PtPin: core.FetchCookieValue(cookie, "pt_pin"),
	// 	}
	// 	type Result struct {
	// 		Code    int         `json:"code"`
	// 		Data    interface{} `json:"data"`
	// 		Message string      `json:"message"`
	// 	}
	// 	result := Result{
	// 		Data: nil,
	// 		Code: 300,
	// 	}
	// 	if ck.PtPin == "" || ck.PtKey == "" {
	// 		result.Message = "一句mmp，不知当讲不当讲。"
	// 		c.JSON(200, result)
	// 		return
	// 	}
	// 	if !ck.Available() {
	// 		result.Message = "无效的ck，请重试。"
	// 		c.JSON(200, result)
	// 		return
	// 	}
	// 	value := fmt.Sprintf(`pt_key=%s;\s*?pt_pin=%s;`, ck.PtKey, ck.PtPin)

	// 	envs, err := GetEnvs(qls[0], "JD_COOKIE")
	// 	if err != nil {
	// 		result.Message = err.Error()
	// 		c.JSON(200, result)
	// 		return
	// 	}
	// 	find := false
	// 	for _, env := range envs {
	// 		if strings.Contains(env.Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) {
	// 			envs = []qinglong.Env{env}
	// 			find = true
	// 			break
	// 		}
	// 	}
	// 	if !find {

	// 		if err := qinglong.AddEnv(qinglong.Env{
	// 			Name:  "JD_COOKIE",
	// 			Value: value,
	// 		}); err != nil {
	// 			result.Message = err.Error()
	// 			c.JSON(200, result)
	// 			return
	// 		}
	// 		rt := ck.Nickname + "，添加成功。"
	// 		core.NotifyMasters(rt)
	// 		result.Message = rt
	// 		result.Code = 200
	// 		c.JSON(200, result)
	// 		return
	// 	} else {
	// 		env := envs[0]
	// 		env.Value = value
	// 		if env.Status != 0 {
	// 			if err := qinglong.Req(nil, qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
	// 				result.Message = err.Error()
	// 				c.JSON(200, result)
	// 				return
	// 			}
	// 			env.Status = 0
	// 			if err := qinglong.UdpEnv(env); err != nil {
	// 				result.Message = err.Error()
	// 				c.JSON(200, result)
	// 				return
	// 			}
	// 		}
	// 		rt := ck.Nickname + "，更新成功。"
	// 		core.NotifyMasters(rt)
	// 		result.Message = rt
	// 		result.Code = 200
	// 		c.JSON(200, result)
	// 		return
	// 	}
	// })
	core.AddCommand("jd", []core.Function{
		// {
		// 	Rules: []string{`unbind ?`},
		// 	Admin: true,
		// 	Handle: func(s core.Sender) interface{} {
		// 		s.Disappear(time.Second * 40)
		// 		envs, err := GetEnvs("JD_COOKIE")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		if len(envs) == 0 {
		// 			return "暂时无法操作。"
		// 		}
		// 		key := s.Get()
		// 		pin := pin(s.GetImType())
		// 		for _, env := range envs {
		// 			pt_pin := FetchJdCookieValue("pt_pin", env.Value)
		// 			pin.Foreach(func(k, v []byte) error {
		// 				if string(k) == pt_pin && string(v) == key {
		// 					s.Reply(fmt.Sprintf("已解绑，%s。", pt_pin))
		// 					pin.Set(string(k), "")
		// 				}
		// 				return nil
		// 			})
		// 		}
		// 		return "操作完成"
		// 	},
		// },
		{
			Rules: []string{"send ? ?"},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				user_pin := s.Get()
				msg := s.Get(1)
				for _, tp := range []string{
					"qq", "tg", "wx",
				} {
					core.Bucket("pin" + strings.ToUpper(tp)).Foreach(func(k, v []byte) error {
						pt_pin := string(k)
						if pt_pin == user_pin || user_pin == "all" {
							if push, ok := core.Pushs[tp]; ok {
								push(string(v), msg, nil, "")
							}
						}
						return nil
					})
				}
				return "发送完成"
			},
		},
		{
			Rules: []string{`unbind`},
			Handle: func(s core.Sender) interface{} {
				s.Disappear(time.Second * 40)

				uid := fmt.Sprint(s.GetUserID())

				pin := pin(s.GetImType())
				pin.Foreach(func(k, v []byte) error {
					if string(v) == uid {
						s.Reply(fmt.Sprintf("已解绑，%s。", string(k)))
						pin.Set(string(k), "")
					}
					return nil
				})
				return "操作完成"
			},
		},
		{
			Rules:   []string{`raw pt_key=([^;=\s]+);\s*pt_pin=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s core.Sender) interface{} {
				if s.GetImType() == "wxsv" && !s.IsAdmin() && jd_cookie.GetBool("ban_wxsv") {
					return "不支持此功能。"
				}
				imType := s.GetImType()
				fake := false
				if strings.HasPrefix(imType, "_") {
					fake = true
					imType = strings.Replace(imType, "_", "", -1)
				}
				if imType == "wxsv" && !s.IsAdmin() {
					return nil
				}
				s.RecallMessage(s.GetMessageID())
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

					qq := ""

					if imType == "qq" {
						qq = s.GetUserID()
					}

					value := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)


					qls := []*qinglong.QingLong{}
					if strings.Contains(jd_cookie.Get("bus"), ck.PtPin) {
						qls = qinglong.GetQLS()
					} else {
						jn := &JdNotify{
							ID: ck.PtPin,
						}
						jdNotify.First(jn)
						err, ql := qinglong.GetQinglongByClientID(jn.ClientID)
						if ql == nil {
							return err.Error()
						}
						qls = []*qinglong.QingLong{ql}
					}

					for _, ql := range qls {
						tail := fmt.Sprintf("	——来自%s", ql.Name)
						if qinglong.GetQLSLen() < 2 {
							tail = ""
						}
						envs, err := GetEnvs(ql, "JD_COOKIE")
						if err != nil {
							s.Reply(err.Error() + tail)
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
						pin(imType).Set(ck.PtPin, s.GetUserID())
						if !find {
							if err := qinglong.AddEnv(ql, qinglong.Env{
								Name:  "JD_COOKIE",
								Value: value,
							}); err != nil {
								s.Reply(err.Error() + tail)
								continue
							}
							rt := ck.Nickname + "，添加成功。"
							core.NotifyMasters(rt + tail)
							s.Reply(rt + tail)
							continue
						} else {
							env := envs[0]
							env.Value = value
							if env.Status != 0 {
								if _, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
									s.Reply(err.Error() + tail)
									continue
								}
							}
							env.Status = 0
							if err := qinglong.UdpEnv(ql, env); err != nil {
								s.Reply(err.Error() + tail)
								continue
							}
							assets.Delete(ck.PtPin)
							rt := ck.Nickname + "，更新成功。"
							core.NotifyMasters(rt + tail)
							s.Reply(rt + tail)
							continue
						}
					}
				}
				return nil
			},
		},
		{
			Rules:   []string{`raw pin=([^;=\s]+);\s*wskey=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s core.Sender) interface{} {
				if s.GetImType() == "wxsv" && !s.IsAdmin() && jd_cookie.GetBool("ban_wxsv") {
					return "不支持此功能。"
				}
				s.Reply(s.Delete())
				s.Disappear(time.Second * 20)
				for _, v := range s.GetAllMatch() {
					value := fmt.Sprintf("pin=%s;wskey=%s;", v[0], v[1])
					pt_key, err := getKey(value)
					if err == nil {
						if strings.Contains(pt_key, "fake") {
							s.Reply("无效的wskey，请重试。")
							continue
						}
					} else {
						s.Reply(err)
					}
					ck := &JdCookie{
						PtKey: pt_key,
						PtPin: v[0],
					}
					ck.Available()

					qls := []*qinglong.QingLong{}
					if strings.Contains(jd_cookie.Get("bus"), ck.PtPin) {
						qls = qinglong.GetQLS()
					} else {
						jn := &JdNotify{
							ID: ck.PtPin,
						}
						jdNotify.First(jn)
						err, ql := qinglong.GetQinglongByClientID(jn.ClientID)
						if ql == nil {
							return err.Error()
						}
						qls = []*qinglong.QingLong{ql}
					}
					for _, ql := range qls {
						tail := fmt.Sprintf("	——来自%s", ql.Name)
						if qinglong.GetQLSLen() < 2 {
							tail = ""
						}
						envs, err := GetEnvs(ql, "pin=")
						if err != nil {
							s.Reply(err.Error() + tail)
							continue
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
							qinglong.AddEnv(ql, qinglong.Env{
								Name:  "JD_COOKIE",
								Value: value2,
							})
						} else {
							if envCK.Status != 0 {
								envCK.Value = value2
								if err := qinglong.UdpEnv(ql, *envCK); err != nil {
									s.Reply(err.Error() + tail)
									continue
								}
							}
						}
						if envWsCK == nil {
							if err := qinglong.AddEnv(ql, qinglong.Env{
								Name:  "JD_WSCK",
								Value: value,
							}); err != nil {
								s.Reply(err.Error() + tail)
								continue
							}
							s.Reply(ck.Nickname + ",添加成功。" + tail)
							continue
						} else {
							envWsCK.Value = value
							if envWsCK.Status != 0 {
								if _, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+envWsCK.ID+`"]`)); err != nil {
									s.Reply(err.Error() + tail)
									continue
								}
							}
							envWsCK.Status = 0
							if err := qinglong.UdpEnv(ql, *envWsCK); err != nil {
								s.Reply(err.Error() + tail)
								continue
							}
							s.Reply(ck.Nickname + ",更新成功。" + tail)
							continue

						}
					}
				}
				return nil
			},
		},
	})
}
