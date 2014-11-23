package main

import "net/smtp"

type Email struct {
	user                string
	pass                string
	notificationAddress string
}

func (a Email) Trigger(event *Event) error {

	body := "To: " + a.notificationAddress +
		"\r\nSubject: " + event.ShortMessage() +
		"\r\n\r\n" + event.ShortMessage()

	auth := smtp.PlainAuth("", a.user, a.pass, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, a.user,
		[]string{a.notificationAddress}, []byte(body))
	if err != nil {
		return err
	}

	event.server.log.Println(white, "Email alert successfully triggered.", reset)
	return nil
}
