package gomail

import (
	"bytes"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"time"
)

// Configuration for mail
type Configuration struct {
	User     string
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Config default configuration
var Config = Configuration{
	Host:     "smtp.qq.com",
	Port:     "25",
	Username: "",
	Password: "",
	From:     "",
}

// GoMail config
type GoMail struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Content string
}

func parseMailAddr(address string) *mail.Address {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		log.Fatalf("Parse email address %s error: %s", address, err)
	}
	return addr
}

// http://tools.ietf.org/html/rfc822
// http://tools.ietf.org/html/rfc2821
func (gm *GoMail) String() string {
	var buf bytes.Buffer
	const crlf = "\r\n"

	write := func(what string, addrs []string) {
		if len(addrs) == 0 {
			return
		}
		for i := range addrs {
			if i == 0 {
				buf.WriteString(what)
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(parseMailAddr(addrs[i]).String())
		}
		buf.WriteString(crlf)
	}

	from := parseMailAddr(gm.From)
	if from.Address == "" {
		from = parseMailAddr(Config.From)
	}
	fmt.Fprintf(&buf, "From: %s%s", from.String(), crlf)
	write("To: ", gm.To)
	write("Cc: ", gm.Cc)
	write("Bcc: ", gm.Bcc)
	fmt.Fprintf(&buf, "Date: %s%s", time.Now().UTC().Format(time.RFC822), crlf)
	fmt.Fprintf(&buf, "Subject: %s%s", gm.Subject, crlf)
	fmt.Fprintf(&buf, "%s%s%s%s", crlf, gm.Content, crlf, crlf)
	return buf.String()
}

// Send email
func (gm *GoMail) Send() error {
	//收件人
	var to []string
	for i := range gm.To {
		to = append(to, parseMailAddr(gm.To[i]).Address)
	}
	//抄送
	for i := range gm.Cc {
		to = append(to, parseMailAddr(gm.Cc[i]).Address)
	}

	fmt.Println(to)

	if gm.From == "" {
		gm.From = Config.From
	}
	from := parseMailAddr(gm.From).Address
	addr := fmt.Sprintf("%s:%s", Config.Host, Config.Port)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Host)
	return smtp.SendMail(addr, auth, from, to, []byte(gm.String()))
}
