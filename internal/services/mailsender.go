package services

import (
	"fmt"
	"net/smtp"

	"github.com/RangoCoder/foodApi/internal/env"
	st "github.com/RangoCoder/foodApi/internal/structs"
)

func SimpleSendEmail(user st.User, mailBody string) (bool, error) {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(env.GetEnvVar("SMTP_SERVER"))
	if err != nil {
		return false, err
	}
	//отправитель
	if err := c.Mail(env.GetEnvVar("SMTP_SERVER_SENDER_MAIL")); err != nil {
		return false, err
	}
	//получатель
	if err := c.Rcpt(user.Email); err != nil {
		return false, err
	}
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return false, err
	}

	_, err = fmt.Fprintf(wc, mailBody)
	if err != nil {
		return false, err
	}
	err = wc.Close()
	if err != nil {
		return false, err
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return false, err
	}
	return true, nil
}
