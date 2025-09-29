package services

import (
	"fmt"
	"log"
	"strconv"

	"github.com/RangoCoder/foodApi/internal/env"
	st "github.com/RangoCoder/foodApi/internal/structs"
	"gopkg.in/gomail.v2"
)

func SimpleSendEmail(user st.User, subject, mailBody string) (bool, error) {

	smtpServer := env.GetEnvVar("SMTP_SERVER")
	smtpPort, err := strconv.Atoi(env.GetEnvVar("SMTP_PORT"))
	if err != nil { //
		return false, err
	}
	smtpSenderMail := env.GetEnvVar("SMTP_SERVER_SENDER_MAIL")
	smtpSenderPass := env.GetEnvVar("SMTP_SERVER_SENDER_PASS")
	d := gomail.NewDialer(smtpServer, smtpPort, smtpSenderMail, smtpSenderPass)
	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "awesome.povel890@yandex.ru")
	m.SetAddressHeader("To", user.Email, "")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", fmt.Sprintf("Письмо сгенерарованно автоматически после регистрации! %v", mailBody))

	if err := gomail.Send(s, m); err != nil {
		log.Printf("Could not send email to %q: %v", user.Email, err)
	}
	m.Reset()

	return true, nil
}

// func SimpleSendEmail(user st.User, mailBody string) (bool, error) {
// 	// Connect to the remote SMTP server.
// 	c, err := smtp.Dial(env.GetEnvVar("SMTP_SERVER"))
// 	if err != nil {
// 		return false, err
// 	}
// 	//отправитель
// 	if err := c.Mail(env.GetEnvVar("SMTP_SERVER_SENDER_MAIL")); err != nil {
// 		return false, err
// 	}
// 	//получатель
// 	if err := c.Rcpt(user.Email); err != nil {
// 		return false, err
// 	}
// 	// Send the email body.
// 	wc, err := c.Data()
// 	if err != nil {
// 		return false, err
// 	}

// 	_, err = fmt.Fprintf(wc, mailBody)
// 	if err != nil {
// 		return false, err
// 	}
// 	err = wc.Close()
// 	if err != nil {
// 		return false, err
// 	}

// 	// Send the QUIT command and close the connection.
// 	err = c.Quit()
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }
