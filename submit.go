package jd_cookie

import (
	"fmt"
	"strings"

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
	core.AddCommand("jd", []core.Function{
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
				if strings.HasPrefix(imType, "_") {
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
	})
}
