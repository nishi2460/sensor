package main

import (
	"fmt"
	"net/smtp"
	"os"
	"time"
)


func send_mail(mainstring string) {

	var subjectstring string

	hostname, _ := os.Hostname()


	tinfo := time.Now()

	subjectstring = "Mime-Version: 1.0\r\n"
	subjectstring += fmt.Sprintf("Subject: %s %02d%02d\r\n", hostname, tinfo.Month(), tinfo.Day())
	subjectstring += fmt.Sprintf("Content-Type: text/plain; charset=iso-2022-jp\r\n")
	subjectstring += fmt.Sprintf("\r\n")
	subjectstring += mainstring + "\r\n"

	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		"sensor.raspi.9831@gmail.com",
		"lnfetzmhnnjsxgwl",
		"smtp.gmail.com",
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	errs := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"sensor.raspi.9831@gmail.com",
		[]string{"nishimura.2460.home@gmail.com"},
		//[]string{"aict.mem2022@gmail.com","Setestse123123@gmail.com","nishimura.2460.home@gmail.com"},
		[]byte(subjectstring),
	)
	if errs != nil {
		fout, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		defer fout.Close()
		res_str := fmt.Sprintf("%v\n", errs)

		fout.WriteString(res_str)

		panic(errs)
	}
}

func main() {
	send_mail("PowerON.2")
}
