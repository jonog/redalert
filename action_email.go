package main

import "net/smtp"

type Email struct {
	user                string
	pass                string
	notificationAddress string
}

func (a Email) Send(server *Server) error {

	body := "To: " + a.notificationAddress + "\r\nSubject: " +
		"Uhoh, " + server.name + " has been nuked!!!" + "\r\n\r\n" +
		"Issue pinging " + server.address
	auth := smtp.PlainAuth("", a.user, a.pass, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, a.user,
		[]string{a.notificationAddress}, []byte(body))
	if err != nil {
		return err
	}
	return nil

}
