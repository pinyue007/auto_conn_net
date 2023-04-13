// 本程序用于公司联网自动登录用
// 编译无窗口程序命令: go install -ldflags -H=windowsgui
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func detect_net() bool {
	//GET一个网页，注意把response关闭掉
	//我使用了golang的一个技能：defer，它能在函数结束前执行指定内容
	res, err := http.Get("http://www.baidu.com")
	if err != nil {
		logrus.Errorln("访问百度出错")
		return false
	}
	defer res.Body.Close()

	//使用ioutil中的函数方便的读取GET到的网页内容。虽然我们也可以自己读取，但哪有这样方便啊。
	data, _ := ioutil.ReadAll(res.Body)
	//查找是否有“未登录”的字符串
	idx := strings.Index(string(data), "<!--STATUS OK-->")
	if idx == -1 {
		logrus.Errorln("当前网络状态异常，即将重新连接！")
		return false
	} else {
		logrus.Infoln("当前网络连接状态正常！")
		return true
	}
}

func reconn_net(user, pwd string) {
	//使用net/url包来管理post数据，对于简单的ASCII内容来说可以简单的自己合成字符串。
	//但使用它，可以保证不会出错。
	v := url.Values{}
	v.Set("param[UserName]", user)
	v.Set("param[UserPswd]", pwd)

	//把form数据编下码
	body := strings.NewReader(v.Encode())

	//建立HTTP对象
	client := &http.Client{}
	//建立http请求对象
	req, _ := http.NewRequest("POST", "http://192.168.255.25/user-login-auth?id=&url=&user=&mac=", body)
	//这个一定要加，不加form的值post不过去
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, _ := client.Do(req) //发送
	if resp == nil {
		logrus.Errorln("client.Do()处理失败！")
		return
	}
	defer resp.Body.Close() //关闭resp.Body
	data, _ := ioutil.ReadAll(resp.Body)
	idx := strings.Index(string(data), "\"msg\":\"success\"")
	if idx != -1 {
		logrus.Infoln("网络重连成功！")
	}
}

func main() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		//以下设置只是为了使输出更美观
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:03:04",
	})
	logPath, _ := os.Getwd()
	logFile := fmt.Sprintf("%s/log.log", logPath)
	w1 := &bytes.Buffer{}
	w2 := os.Stdout
	w3, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("创建日志文件失败: %v", err)
	}
	logrus.SetOutput(io.MultiWriter(w1, w2, w3))

	//下面三句用于处理参数，注意flag.Parse()必须调用。
	//flag包非常实用，可以自动处理命令行参数的很多细节。
	//比如：自动处理--help，它会输出我们设定的信息
	user := flag.String("user", "", "input your user name.")
	pwd := flag.String("pwd", "", "input your password.")
	sec := flag.Int("sleep", 10, "input sleep minites, default sleep 30 minites.")

	flag.Parse()
	if *user == "" || *pwd == "" {
		logrus.Warnln("please input your user name and password!")
		logrus.Warnln("Usage: net.exe -user=xxxxxx -pwd=yyyyyy -sleep=30")
		return
	}

	if *sec == 0 {
		*sec = 30
	}

	// 死循环，每隔一段时间检查网络，网络异常重连
	for {
		logrus.Infoln("准备检测网络……")
		net_status := detect_net()
		if !net_status {
			reconn_net(*user, *pwd)
		}
		time.Sleep(time.Duration(*sec) * time.Second * 60) //默认睡眠30分钟
		// time.Sleep(time.Second * 10) //调试，睡眠10秒
	}
}
