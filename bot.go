package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/tidwall/gjson"
)

//https://developers.line.biz/
const channelSecret = "YOUR_CHANNEL_SECRET"
const channelToken = "YOUR_CHANNEL_TOKEN"

//https://openweathermap.org/api
const weatherKey = "YOUR_WEATHER_KEY"

func main() {
	fmt.Println("start...")
	http.HandleFunc("/", BotServer)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err.Error())
	}
}

func BotServer(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Connect success")
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		fmt.Println(err.Error())
	}
	events, err := bot.ParseRequest(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					fmt.Print(err)
				}
			case *linebot.LocationMessage:
				lat := strconv.FormatFloat(message.Latitude, 'f', 2, 64)
				lon := strconv.FormatFloat(message.Longitude, 'f', 2, 64)
				currentWeatherMessage := linebot.NewTextMessage(currentWeather(lat, lon))
				forecastWeatherMessage := linebot.NewTextMessage(forecastWeather(lat, lon))
				msg := currentWeatherMessage.Text + forecastWeatherMessage.Text
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
					fmt.Print(err)
				}
			}
		}
	}
}

func currentWeather(lat, lon string) string {
	resp, _ := http.Get("https://api.openweathermap.org/data/2.5/weather?lat=" + lat + "&lon=" + lon + "&appid=" + weatherKey)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	tm1 := time.Unix(time.Now().Unix(), 0)
	date := gjson.Get(string(body), "sys.country").String()
	date = date + " " + gjson.Get(string(body), "name").String() + " " + tm1.Format("01/02 15:04")
	tempF := gjson.Get(string(body), "main.temp_max").Float() - 273.15
	tempS := strconv.FormatFloat(tempF, 'f', 2, 64)
	date = date + "\ntemp: " + tempS + "°C  " + gjson.Get(string(body), "weather.0.description").String()

	return date
}

func forecastWeather(lat, lon string) string {
	resp, _ := http.Get("https://api.openweathermap.org/data/2.5/forecast?lat=" + lat + "&lon=" + lon + "&appid=" + weatherKey)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	indexDate := time.Unix(gjson.Get(string(body), "list.0.dt").Int(), 0).Format("01/02")
	currentDate := ""
	currentTime := ""
	currentTemp := ""
	date := "\n***********" + indexDate + "***********"
	for i := 0; i < int(gjson.Get(string(body), "cnt").Int()); i++ {
		currentDate = time.Unix(gjson.Get(string(body), "list."+strconv.Itoa(i)+".dt").Int(), 0).Format("01/02")
		if indexDate != currentDate {
			date += "\n***********" + currentDate + "***********"
			indexDate = currentDate
		}
		currentTime = time.Unix(gjson.Get(string(body), "list."+strconv.Itoa(i)+".dt").Int(), 0).Format("15:04")
		temp := gjson.Get(string(body), "list."+strconv.Itoa(i)+".main.temp_max").Float() - 273.15
		currentTemp = strconv.FormatFloat(temp, 'f', 2, 64)
		date = date + "\n" + currentTime + " " + currentTemp + "°C"
		date = date + " " + gjson.Get(string(body), "list."+strconv.Itoa(i)+".weather.0.description").String()
	}

	return date
}
