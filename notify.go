package jd_cookie

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
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
				head += "\n⚠️⚠️⚠️账号即将过期，请更新CK。\n\n"
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
			"提醒：该账号已过期，请发送账号信息。”",
		}...)
	}
	ck.PtPin, _ = url.QueryUnescape(ck.PtPin)
	rt := strings.Join(msgs, "\n")
	if jd_cookie.GetBool("tuyalize", false) == true {

	}
	return rt
}
