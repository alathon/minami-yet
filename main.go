package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/spf13/viper"
)

type app struct {
	lastBodyHash     []byte
	lastCheckTime    time.Time
	registeredEmails []string
}

func sendEmail(a *app) {
	from := viper.GetString("SMTP_FROM")
	password := viper.GetString("SMTP_PASSWORD")
	host := viper.GetString("SMTP_HOST")
	port := viper.GetString("SMTP_PORT")
	addr := host + ":" + port
	to := a.registeredEmails

	subject := "Subject: Change on website?\n"
	body := "The website content has changed"
	message := []byte(subject + body)

	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(addr, auth, from, to, message)
	if err != nil {
		panic(err)
	}
}

func main() {
	viper.SetEnvPrefix("vpr")
	viper.AutomaticEnv()
	viper.SetDefault("TIME_BETWEEN_SECONDS", 3600)

	timeBetweenSeconds := viper.GetInt("TIME_BETWEEN_SECONDS")
	website := viper.GetString("WEBSITE")

	a := app{
		lastCheckTime:    time.Now(),
		registeredEmails: strings.Split(viper.GetString("EMAILS"), ","),
	}

	coll := colly.NewCollector()
	coll.AllowURLRevisit = true

	coll.OnResponse(func(r *colly.Response) {
		h := sha1.New()
		h.Write(r.Body)
		hByte := h.Sum(nil)
		if !bytes.Equal(a.lastBodyHash, hByte) {
			fmt.Println("New body hash: ", hex.EncodeToString(hByte))
			a.lastBodyHash = hByte
			sendEmail(&a)
		}
	})

	for {
		coll.Visit(website)
		time.Sleep(time.Duration(timeBetweenSeconds) * time.Second)
	}

}
