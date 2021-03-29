package main

import (
	"Email/gomail"
	"encoding/json"
	"fmt"
	env "github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

//Receiver for receive email
type Receiver struct {
	Email string `json:"email"`
	Local string `json:"local"`
}

func isDev() bool {
	return os.Getenv("MAIL_MODE") == "dev"
}

func main() {
	loadConfig()

	if isDev() {
		batchSendMail()
		return
	}

	nyc, _ := time.LoadLocation("Asia/Shanghai")
	cJob := cron.New(cron.WithLocation(nyc))

	cronCfg := os.Getenv("MAIL_CRON")
	if cronCfg == "" {
		batchSendMail()
	} else {
		cJob.AddFunc(cronCfg, func() {
			batchSendMail()
		})
		cJob.Start()
		select {}
	}
}

func loadConfig() {
	err := env.Load()
	if err != nil {
		log.Fatalf("Load .env file error: %s", err)
	}
}

func batchSendMail() {
	receivers := getReceivers("MAIL_TO")
	if len(receivers) == 0 {
		return
	}
	Ccs := getCcs("MAIL_CC")

	wg := sync.WaitGroup{}
	for _, receiver := range receivers {
		wg.Add(1)
		go func(receiver Receiver) {
			defer wg.Done()
			sendMail(receiver.Email, Ccs)
		}(receiver)
	}
	wg.Wait()
}

func getReceivers(encReceiver string) []Receiver {
	var receivers []Receiver
	userJSON := os.Getenv(encReceiver)
	err := json.Unmarshal([]byte(userJSON), &receivers)
	if err != nil {
		log.Fatalf("Parse users from %s error: %s", userJSON, err)
	}
	return receivers
}

func getCcs(envCc string) []Receiver {
	var Ccs []Receiver
	userJSON := os.Getenv(envCc)
	err := json.Unmarshal([]byte(userJSON), &Ccs)
	if err != nil {
		log.Fatalf("Parse users from %s error: %s", userJSON, err)
	}
	return Ccs
}

func sendMail(to string, Ccs []Receiver) {
	gomail.Config.User = os.Getenv("MAIL_USER")
	gomail.Config.Username = os.Getenv("MAIL_USERNAME")
	gomail.Config.Password = os.Getenv("MAIL_PASSWORD")
	gomail.Config.Host = os.Getenv("MAIL_HOST")
	gomail.Config.Port = os.Getenv("MAIL_PORT")
	gomail.Config.From = os.Getenv("MAIL_FROM")

	today := time.Now().Format("2006.1.2")
	b, err := ioutil.ReadFile(today)
	if err != nil {
		log.Fatal(err.Error())
	}
	duration := time.Now().AddDate(0, 0, 6).Format("2006.1.2")
	week := today + "-" + duration

	var ccs []string
	for _, cc := range Ccs {
		ccs = append(ccs, cc.Email)
	}

	subject := fmt.Sprintf("%s %s%s", gomail.Config.User, week, os.Getenv("MAIL_SUBJECT"))

	email := gomail.GoMail{
		To:      []string{to},
		Subject: subject,
		Cc:      ccs,
		Content: string(b),
	}

	err = email.Send()
	if err != nil {
		log.Printf("Send email fail, error: %s", err)
	} else {
		log.Printf("Send email %s success!", to)
	}
}
