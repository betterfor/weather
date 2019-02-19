package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

const (
	url  = "https://restapi.amap.com/v3/weather/weatherInfo?"
	key  = "fcb0e944f85f0eb0ef11974f45ccc3ba"
	city = "370211" // 城市编码
	ext  = "all"    // all返回预报天气，base返回实况天气
)

type Weather struct {
	Status    string     `json:"status"返回状态`
	Count     string     `json:"count"返回结果总条数`
	Info      string     `json:"info"返回的状态信息`
	Infocode  string     `json:"infocode"返回状态说明`
	Forecasts []Forecast `json:"forecasts"预报天气信息数据`
}
type Forecast struct {
	City       string `json:"city"城市名称`
	Adcode     string `json:"adcode"城市编码`
	Province   string `json:"province"省份`
	Reporttime string `json:"reporttime"预报时间`
	Casts      []Cast `json:casts预报数据`
}
type Cast struct {
	Date         string `json:"date"日期`
	Week         string `json:"week"星期`
	Dayweather   string `json:"dayweather"白天天气`
	Nightweather string `json:"nightweather"晚上天气`
	Daytemp      string `json:"daytemp"白天温度`
	Nighttemp    string `json:"nighttemp"晚上温度`
	Daywind      string `json:"daywind"白天风向`
	Nightwind    string `json:"nightwind"晚上风向`
	Daypower     string `json:"daypower"白天风力`
	Nightpower   string `json:"nightpower"晚上风力`
}

// 网络请求
func doHttpGetRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	return string(body), nil
}

// 获取天气信息
func getWeather() (string, string, error) {
	var data Weather
	var fore Forecast
	var cast Cast
	var str string
	urlInfo := url + "city=" + city + "&key=" + key + "&extensions=" + ext
	rlt, err := doHttpGetRequest(urlInfo)
	if err != nil {
		return "网络访问失败", "", err
	}

	err = json.Unmarshal([]byte(rlt), &data)
	if err != nil {
		return "json数据解析失败", "", err
	}
	fore = data.Forecasts[0]
	output := fore.Province + fore.City + "预报时间：" + fore.Reporttime + "\n"
	for i := 0; i < len(fore.Casts); i++ {
		cast = fore.Casts[i]
		str += "日期" + cast.Date + "\t星期" + NumToStr(cast.Week) +
			"\n白天：【天气：" + cast.Dayweather + "\t温度：" + cast.Daytemp + "\t风向" + cast.Daywind + "\t风力：" + cast.Daypower + "】" +
			"\n夜晚：【天气：" + cast.Nightweather + "\t温度：" + cast.Nighttemp + "\t风向" + cast.Nightwind + "\t风力：" + cast.Nightpower + "】"
	}
	subject := verify(fore.Casts[0].Dayweather, fore.Casts[0].Nightweather)
	return subject, output + str, nil
}

func NumToStr(str string) string {
	switch str {
	case "1":
		return "一"
	case "2":
		return "二"
	case "3":
		return "三"
	case "4":
		return "四"
	case "5":
		return "五"
	case "6":
		return "六"
	case "7":
		return "日"
	default:
		return ""
	}
}

func verify(dayweather, nightweather string) string {
	var sub string
	rain := "雨"
	snow := "雪"
	sub = "今日天气预报"
	if strings.Contains(dayweather, rain) || strings.Contains(nightweather, rain) {
		sub = sub + "今天将降雨，出门请别忘带伞"
	}
	if strings.Contains(dayweather, snow) || strings.Contains(nightweather, snow) {
		sub = sub + "    下雪了"
	}
	return sub
}

// 发送邮件
func sendToMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])

	var content_type string
	if mailtype == "html" {
		content_type = "Content_Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content_Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To:" + to + "\r\nFrom: " + user + "<" +
		user + ">\r\nSubject: " + subject + "\r\n" +
		content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

//
func sendEmail(subject, body string) {
	user := "发件箱"
	pwd := "发件箱的授权码"
	host := "smtp.126.com:25"
	to := "收件箱" //可以用;隔开发送多个
	fmt.Println("send email")
	err := sendToMail(user, pwd, host, to, subject, body, "html")
	if err != nil {
		fmt.Println("Send mail error!")
		fmt.Println(err)
	} else {
		fmt.Println("Send mail success!")
	}
}

// 定时结算（一天发一次）
func TimeSettle() {
	d := time.Duration(time.Minute)
	t := time.NewTicker(d)
	defer t.Stop()
	for {
		currentTime := time.Now()
		if currentTime.Hour() == 8 { // 8点发送
			sendinfo()
			time.Sleep(time.Hour)
		}
		<-t.C
	}
}

func sendinfo() {
	subject, body, err := getWeather()
	if err != nil {
		fmt.Println(err)
	}
	sendEmail(subject, body)
}

func main() {
	TimeSettle()
	//sendinfo()
}
