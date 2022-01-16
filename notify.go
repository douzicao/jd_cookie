package jd_cookie

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
	"github.com/douzicao/sillyGirl/core"
	"github.com/douzicao/sillyGirl/develop/qinglong"
	cron "github.com/robfig/cron/v3"
)

type JdNotify struct {
	ID           string
	Pet          bool
	Fruit        bool
	DreamFactory bool
	Note         string
	PtKey        string
	AssetCron    string
	PushPlus     string
	LoginedAt    time.Time
	ClientID     string
}

var cc *cron.Cron

var jdNotify = core.NewBucket("jdNotify")

func assetPush(pt_pin string) {
	jn := &JdNotify{
		ID: pt_pin,
	}
	jdNotify.First(jn)
	if jn.PushPlus != "" {
		// tail := ""
		head := ""

		days, hours, minutes, seconds := getDifference(jn.LoginedAt, time.Now())
		if days < 1000 {
			head = fmt.Sprintf("登录时长：%d天%d时%d分%d秒", days, hours, minutes, seconds)
			if days > 25 {
				head += "\n⚠️⚠️⚠️账号即将过期，请登录。\n\n"
			} else {
				head += "\n\n"
			}
		}

		pushpluspush("资产变动通知", head+GetAsset(&JdCookie{
			PtPin: pt_pin,
			PtKey: jn.PtKey,
		}), jn.PushPlus)
		return
	}
	qqGroup := jd_cookie.GetInt("qqGroup")
	if jn.PtKey != "" && pt_pin != "" {
		pt_key := jn.PtKey
		for _, tp := range []string{
			"qq", "tg", "wx",
		} {
			var fs []func()
			core.Bucket("pin" + strings.ToUpper(tp)).Foreach(func(k, v []byte) error {
				if string(k) == pt_pin && pt_pin != "" {
					if push, ok := core.Pushs[tp]; ok {
						fs = append(fs, func() {
							push(string(v), GetAsset(&JdCookie{
								PtPin: pt_pin,
								PtKey: pt_key,
							}), qqGroup, "")
						})
					}
				}
				return nil
			})
			if len(fs) != 0 {
				for _, f := range fs {
					f()
				}
			}
		}
	}
}

var ccc = map[string]cron.EntryID{}

func initNotify() {
	cc = cron.New(cron.WithSeconds())
	cc.Start()
	jdNotify.Foreach(func(_, v []byte) error {
		aa := &JdNotify{}
		json.Unmarshal(v, aa)
		if aa.AssetCron != "" {
			if rid, err := cc.AddFunc(aa.AssetCron, func() {
				assetPush(aa.ID)
			}); err == nil {
				ccc[aa.ID] = rid
			}
		}
		return nil
	})
	go func() {
		time.Sleep(time.Second)
		for {
			for _, ql := range qinglong.GetQLS() {
				as := 0
				envs, _ := GetEnvs(ql, "JD_COOKIE")
				for _, env := range envs {

					if env.Status != 0 {
						continue
					}
					as++
					pt_pin := core.FetchCookieValue(env.Value, "pt_pin")
					pt_key := core.FetchCookieValue(env.Value, "pt_key")
					if pt_pin != "" && pt_key != "" {
						jn := &JdNotify{
							ID: pt_pin,
						}
						jdNotify.First(jn)
						tc := false
						if jn.PtKey != pt_key {
							jn.PtKey = pt_key
							tc = true
						}
						if jn.ClientID != ql.ClientID {
							jn.ClientID = ql.ClientID
							tc = true
						}
						if tc {
							jdNotify.Create(jn)
						}
					}
				}
				ql.SetNumber(as)
			}
			time.Sleep(time.Second * 30)
		}
	}()
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^任务通知$`},
			Cron:  jd_cookie.Get("task_Notify", "2 7,13,19 * * *"),
			Admin: true,
			Handle: func(_ core.Sender) interface{} {
				jdNotify.Foreach(func(_, v []byte) error {
					aa := &JdNotify{}
					if json.Unmarshal(v, aa) == nil {
						ck := fmt.Sprintf("pt_key=%s;pt_pin=%s;", aa.PtKey, aa.ID)
						initPetTown(ck, nil)
						initFarm(ck, nil)
						dream(ck, nil)
					}
					return nil
				})
				return "推送完成"
			},
		},
		{
			Rules: []string{`raw ^关闭(.+)通知$`},
			Handle: func(s core.Sender) interface{} {
				class := s.Get()
				pin := pin(s.GetImType())
				uid := fmt.Sprint(s.GetUserID())
				accounts := []string{}
				pin.Foreach(func(k, v []byte) error {
					if string(v) == uid {
						accounts = append(accounts, string(k))
					}
					return nil
				})
				for i := range accounts {
					jn := &JdNotify{
						ID: accounts[i],
					}
					jdNotify.First(jn)
					if class == "京喜工厂" {
						jn.DreamFactory = true
					}
					if class == "东东农场" {
						jn.Fruit = true
					}
					if class == "东东萌宠" {
						jn.Pet = true
					}
					jdNotify.Create(jn)
				}
				return fmt.Sprintf("已为你关闭%d个账号的"+class+"通知。", len(accounts))
			},
		},
		{
			Rules: []string{`raw ^账号管理$`},
			Handle: func(s core.Sender) interface{} {
				if groupCode := jd_cookie.Get("groupCode"); !s.IsAdmin() && groupCode != "" && s.GetChatID() != 0 && !strings.Contains(groupCode, fmt.Sprint(s.GetChatID())) {
					s.Continue()
					return nil
				}
				pin := pin(s.GetImType())
				uid := fmt.Sprint(s.GetUserID())
				accounts := []string{}
				pin.Foreach(func(k, v []byte) error {
					if string(v) == uid {
						accounts = append(accounts, string(k))
					}
					return nil
				})
				num := len(accounts)
				if num == 0 {
					return "抱歉，你还没有绑定的账号呢~"
				}
				ask := fmt.Sprintf("请在20秒内从1~%d中选择你要操作的账号：\n", num)
				for i := range accounts {
					jn := &JdNotify{
						ID: accounts[i],
					}
					jdNotify.First(jn)
					note := ""
					if jn.Note != "" {
						note = jn.Note
					} else {
						note = jn.ID
					}
					ask += fmt.Sprintf("%d. %s\n", i+1, note)
				}
				s.Reply(strings.Trim(ask, "\n"))
				rt := s.Await(s, func(s core.Sender) interface{} {
					return core.Range([]int{1, num})
				}, time.Second*20)
				switch rt.(type) {
				case nil:
					return "超时，已退出会话。"
				case int:
					pt_pin := accounts[rt.(int)-1]
					jn := &JdNotify{
						ID: pt_pin,
					}
					jdNotify.First(jn)
					ask := "请在20秒内选择操作：\n1. 推送账号资产\n"

					if jn.Note == "" {
						ask += "2. 添加账户备注信息\n"
					} else {
						ask += "2. 修改账户备注信息\n"
					}
					if jn.Pet {
						ask += "3. 开启东东萌宠通知\n"
					} else {
						ask += "3. 关闭东东萌宠通知\n"
					}
					if jn.Fruit {
						ask += "4. 开启东东果园通知\n"
					} else {
						ask += "4. 关闭东东果园通知\n"
					}
					if jn.DreamFactory {
						ask += "5. 开启京喜工厂通知\n"
					} else {
						ask += "5. 关闭京喜工厂通知\n"
					}
					if jn.AssetCron == "" {
						ask += "6. 添加资产推送时间\n"
					} else {
						ask += "6. 修改资产推送时间\n"
					}
					ask += "7. 解绑当前账号\n8. 设置微信push+通知(推荐)\n9. 退出当前会话"
					s.Reply(ask)
					rt := s.Await(s, func(s core.Sender) interface{} {
						return core.Range([]int{1, 9})
					}, time.Second*20)
					switch rt.(type) {
					case nil:
						return "超时，已退出会话。"
					case int:
						switch rt.(int) {
						case 1:
							if jn.PtKey == "" {
								return "账号已过期，暂时无法查询。"
							}
							assetPush(jn.ID)
							return "推送完成，请查收。"
						case 2:
							s.Reply("请输入新的账号备注信息：")
							jn.Note = s.Await(s, nil).(string)
						case 3:
							jn.Pet = !jn.Pet
						case 4:
							jn.Fruit = !jn.Fruit
						case 5:
							jn.DreamFactory = !jn.DreamFactory
						case 6:
							s.Reply("请输入资产推送时间(格式00:00:00，对应时、分、秒):")
							rt := s.Await(s, nil).(string)
							_, err := time.ParseInLocation("2006-01-02 15:04:05", time.Now().Format("2006-01-02"+" ")+rt, time.Local)
							if err != nil {
								s.Reply("格式错误，已退出会话。")
								return nil
							}
							dd := strings.Split(rt, ":")
							jn.AssetCron = fmt.Sprintf("%s %s %s * * *", dd[2], dd[1], dd[0])
							if rid, ok := ccc[jn.ID]; ok {
								cc.Remove(rid)
								if rid, err := cc.AddFunc(jn.AssetCron, func() {
									assetPush(jn.ID)
								}); err == nil {
									ccc[jn.ID] = rid
								} else {
									return err
								}
							}
						case 7:
							pin.Set(pt_pin, "")
							return "解绑成功，会话结束。"
						case 8: //欢迎叼毛看内裤
							data, _ := httplib.Get("https://www.pushplus.plus/api/common/wechat/getQrcode").Bytes()
							qrCodeUrl, _ := jsonparser.GetString(data, "data", "qrCodeUrl")
							qrCode, _ := jsonparser.GetString(data, "data", "qrCode")
							if qrCodeUrl == "" {
								return "嗝屁了"
							}
							s.Reply("请在30秒内打开微信扫描二维码关注公众号：\n" + core.ToImage(qrCodeUrl))
							ck := ""
							n := time.Now()
							for {
								if n.Add(time.Second * 30).Before(time.Now()) {
									return "扫码超时。"
								}
								time.Sleep(time.Second)
								rsp, err := httplib.Get("https://www.pushplus.plus/api/common/wechat/confirmLogin?key=" + qrCode + "&code=1001").Response()
								if err != nil {
									continue
								}
								ck = rsp.Header.Get("Set-Cookie")
								if ck != "" {
									// fmt.Println(ck)
									break
								}
							}
							req := httplib.Get("https://www.pushplus.plus/api/customer/user/token")
							req.Header("Cookie", ck)
							data, _ = req.Bytes()
							jn.PushPlus, _ = jsonparser.GetString(data, "data")
							s.Reply("扫码成功，将尝试为你推送资产信息。")
							assetPush(jn.ID)
						case 9:
							return "已退出会话。"
						}
					}
					jdNotify.Create(jn)
					return "操作成功，会话结束。"
				}
				return nil
			},
		},
	})
}

func a叉哦叉哦(pt_pin, class, content string) {
	u := &JdNotify{
		ID: pt_pin,
	}
	jdNotify.First(u)
	if u.DreamFactory && class == "京喜工厂" {
		return
	}
	if u.Fruit && class == "东东农场" {
		return
	}
	if u.Pet && class == "东东萌宠" {
		return
	}
	if u.Note == "" {
		u.Note = u.ID
	}
	u.Note, _ = url.QueryUnescape(u.Note)
	if u.PushPlus != "" {
		pushpluspush(class+"通知", content+"\n\n通知没有用？请对登录机器人说“关闭"+class+"通知”或“账号管理”，根据提示进行关闭。", u.PushPlus)
		return
	}
	Notify(pt_pin, class+"通知("+u.Note+")：\n"+content+"\n\n通知没有用？请对我说“关闭"+class+"通知”或“账号管理”，根据提示进行关闭。")
}

func pushpluspush(title, content, token string) {
	req := httplib.Post("http://www.pushplus.plus/send")
	req.JSONBody(map[string]string{
		"token":    token,
		"title":    title,
		"content":  content,
		"template": "txt",
	})
	req.Response()
}

func (ck *JdCookie) QueryAsset() string {
	msgs := []string{}
	if ck.Note != "" {
		msgs = append(msgs, fmt.Sprintf("账号备注：%s", ck.Note))
	}
	asset := Asset{}
	if ck.Available() {
		// msgs = append(msgs, fmt.Sprintf("用户等级：%v", ck.UserLevel))
		// msgs = append(msgs, fmt.Sprintf("等级名称：%v", ck.LevelName))
		cookie := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
		var rpc = make(chan []RedList)
		var fruit = make(chan string)
		var pet = make(chan string)
		var dm = make(chan string)
		var gold = make(chan int64)
		var egg = make(chan int64)
		var tyt = make(chan string)
		var mmc = make(chan int64)
		var zjb = make(chan int64)
		var xdm = make(chan []int)
		// var jxz = make(chan string)
		var jrjt = make(chan string)
		var sysp = make(chan string)
		var wwjf = make(chan int)
		// go jingxiangzhi(cookie, jxz)
		go queryuserjingdoudetail(cookie, xdm)
		go dream(cookie, dm)
		go redPacket(cookie, rpc)
		go initFarm(cookie, fruit)
		go initPetTown(cookie, pet)
		go jsGold(cookie, gold)
		go jxncEgg(cookie, egg)
		go tytCoupon(cookie, tyt)
		go mmCoin(cookie, mmc)
		go jdzz(cookie, zjb)
		go jingtie(cookie, jrjt)
		go jdsy(cookie, sysp)
		go cwwjf(cookie, wwjf)

		today := time.Now().Local().Format("2006-01-02")
		yestoday := time.Now().Local().Add(-time.Hour * 24).Format("2006-01-02")
		page := 1
		end := false
		var xdd []int
		for {
			if end {
				xdd = <-xdm
				ti := []string{}
				if asset.Bean.YestodayIn != 0 {
					ti = append(ti, fmt.Sprintf("%d京豆", asset.Bean.YestodayIn))
				}
				if xdd[3] != 0 {
					ti = append(ti, fmt.Sprintf("%d喜豆", xdd[3]))
				}
				if len(ti) > 0 {
					msgs = append(msgs,
						"昨日收入："+strings.Join(ti, "、"),
					)
				}
				ti = []string{}
				if asset.Bean.YestodayOut != 0 {
					ti = append(ti, fmt.Sprintf("%d京豆", asset.Bean.YestodayOut))
				}
				if xdd[4] != 0 {
					ti = append(ti, fmt.Sprintf("%d喜豆", xdd[4]))
				}
				if len(ti) > 0 {
					msgs = append(msgs,
						"昨日支出："+strings.Join(ti, "、"),
					)
				}
				ti = []string{}
				if asset.Bean.TodayIn != 0 {
					ti = append(ti, fmt.Sprintf("%d京豆", asset.Bean.TodayIn))
				}
				if xdd[1] != 0 {
					ti = append(ti, fmt.Sprintf("%d喜豆", xdd[1]))
				}
				if len(ti) > 0 {
					msgs = append(msgs,
						"今日收入："+strings.Join(ti, "、"),
					)
				}
				ti = []string{}
				if asset.Bean.TodayOut != 0 {
					ti = append(ti, fmt.Sprintf("%d京豆", asset.Bean.TodayOut))
				}
				if xdd[2] != 0 {
					ti = append(ti, fmt.Sprintf("%d喜豆", xdd[2]))
				}
				if len(ti) > 0 {
					msgs = append(msgs,
						"今日支出："+strings.Join(ti, "、"),
					)
				}
				break
			}
			bds := getJingBeanBalanceDetail(page, cookie)
			if bds == nil {
				end = true
				msgs = append(msgs, "京豆数据异常")
				break
			}
			for _, bd := range bds {
				amount := Int(bd.Amount)
				if strings.Contains(bd.Date, today) {
					if amount > 0 {
						asset.Bean.TodayIn += amount
					} else {
						asset.Bean.TodayOut += -amount
					}
				} else if strings.Contains(bd.Date, yestoday) {
					if amount > 0 {
						asset.Bean.YestodayIn += amount
					} else {
						asset.Bean.YestodayOut += -amount
					}
				} else {
					end = true
					break
				}
			}
			page++
		}
		var ti []string
		if ck.BeanNum != "" {
			ti = append(ti, ck.BeanNum+"京豆")
		}
		if len(xdd) > 0 && xdd[0] != 0 {
			ti = append(ti, fmt.Sprint(xdd[0])+"喜豆")
		}
		if len(ti) > 0 {
			msgs = append(msgs, "当前豆豆："+strings.Join(ti, "、"))
		}
		ysd := int(time.Now().Add(24 * time.Hour).Unix())
		if rps := <-rpc; len(rps) != 0 {
			for _, rp := range rps {
				b := Float64(rp.Balance)
				asset.RedPacket.Total += b
				if strings.Contains(rp.ActivityName, "京喜") || strings.Contains(rp.OrgLimitStr, "京喜") {
					asset.RedPacket.Jx += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJx += b
						asset.RedPacket.ToExpire += b
					}
				} else if strings.Contains(rp.ActivityName, "极速版") {
					asset.RedPacket.Js += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJs += b
						asset.RedPacket.ToExpire += b
					}

				} else if strings.Contains(rp.ActivityName, "京东健康") {
					asset.RedPacket.Jk += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJk += b
						asset.RedPacket.ToExpire += b
					}
				} else {
					asset.RedPacket.Jd += b
					if ysd >= rp.EndTime {
						asset.RedPacket.ToExpireJd += b
						asset.RedPacket.ToExpire += b
					}
				}
			}
			e := func(m float64) string {
				if m > 0 {
					return fmt.Sprintf(`(今日过期%.2f)`, m)
				}
				return ""
			}
			if asset.RedPacket.Total != 0 {
				msgs = append(msgs, fmt.Sprintf("所有红包：%.2f%s元🧧", asset.RedPacket.Total, e(asset.RedPacket.ToExpire)))
				if asset.RedPacket.Jx != 0 {
					msgs = append(msgs, fmt.Sprintf("京喜红包：%.2f%s元", asset.RedPacket.Jx, e(asset.RedPacket.ToExpireJx)))
				}
				if asset.RedPacket.Js != 0 {
					msgs = append(msgs, fmt.Sprintf("极速红包：%.2f%s元", asset.RedPacket.Js, e(asset.RedPacket.ToExpireJs)))
				}
				if asset.RedPacket.Jd != 0 {
					msgs = append(msgs, fmt.Sprintf("京东红包：%.2f%s元", asset.RedPacket.Jd, e(asset.RedPacket.ToExpireJd)))
				}
				if asset.RedPacket.Jk != 0 {
					msgs = append(msgs, fmt.Sprintf("健康红包：%.2f%s元", asset.RedPacket.Jk, e(asset.RedPacket.ToExpireJk)))
				}
			}

		} else {
			// msgs = append(msgs, "暂无红包数据🧧")
		}
		msgs = append(msgs, fmt.Sprintf("东东农场：%s", <-fruit))
		msgs = append(msgs, fmt.Sprintf("东东萌宠：%s", <-pet))

		msgs = append(msgs, fmt.Sprintf("京东试用：%s", <-sysp))

		msgs = append(msgs, fmt.Sprintf("金融金贴：%s元💰", <-jrjt))

		gn := <-gold
		// if gn >= 30000 {
		msgs = append(msgs, fmt.Sprintf("极速金币：%d(≈%.2f元)💰", gn, float64(gn)/10000))
		// }
		zjbn := <-zjb
		// if zjbn >= 50000 {
		msgs = append(msgs, fmt.Sprintf("京东赚赚：%d金币(≈%.2f元)💰", zjbn, float64(zjbn)/10000))
		// } else {
		// msgs = append(msgs, fmt.Sprintf("京东赚赚：暂无数据"))
		// }
		mmcCoin := <-mmc
		// if mmcCoin >= 3000 {
		msgs = append(msgs, fmt.Sprintf("京东秒杀：%d秒秒币(≈%.2f元)💰", mmcCoin, float64(mmcCoin)/1000))
		// } else {
		// msgs = append(msgs, fmt.Sprintf("京东秒杀：暂无数据"))
		// }

		msgs = append(msgs, fmt.Sprintf("汪汪积分：%d积分", <-wwjf))
		msgs = append(msgs, fmt.Sprintf("京喜工厂：%s", <-dm))
		// if tyt := ; tyt != "" {
		msgs = append(msgs, fmt.Sprintf("推一推券：%s", <-tyt))
		// }
		// if egg := ; egg != 0 {
		msgs = append(msgs, fmt.Sprintf("惊喜牧场：%d枚鸡蛋🥚", <-egg))
		// }
		// if ck.Note != "" {
		// 	msgs = append([]string{
		// 		fmt.Sprintf("账号备注：%s", ck.Note),
		// 	}, msgs...)
		// }
		if runtime.GOOS != "darwin" {
			if ck.Nickname != "" {
				msgs = append([]string{
					fmt.Sprintf("账号昵称：%s", ck.Nickname),
				}, msgs...)
			}
		}
	} else {
		ck.PtPin, _ = url.QueryUnescape(ck.PtPin)
		msgs = append(msgs, fmt.Sprintf("京东账号：%s", ck.PtPin))
		msgs = append(msgs, []string{
			// "提醒：该账号已过期，请重新登录。多账号的🐑毛党员注意了，登录第2个账号的时候，不可以退出第1个账号，退出会造成账号过期。可以在登录第2个账号前清除浏览器cookie，或者使用浏览器的无痕模式。",
			"提醒：该账号已过期，请对我说“登录“。”",
		}...)
	}
	ck.PtPin, _ = url.QueryUnescape(ck.PtPin)
	rt := strings.Join(msgs, "\n")
	if jd_cookie.GetBool("tuyalize", false) == true {

	}
	return rt
}
