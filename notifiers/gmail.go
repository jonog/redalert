package notifiers

import (
	"errors"
	"net/smtp"
	"strings"
)

func init() {
	registerNotifier("gmail", NewGmailNotifier)
}

type Gmail struct {
	name                  string
	user                  string
	pass                  string
	notificationAddresses []string
}

var NewGmailNotifier = func(config Config) (Notifier, error) {

	if config.Type != "gmail" {
		return nil, errors.New("gmail: invalid config type")
	}

	if config.Config["user"] == "" {
		return nil, errors.New("gmail: invalid user")
	}

	if config.Config["pass"] == "" {
		return nil, errors.New("gmail: invalid pass")
	}

	if config.Config["notification_addresses"] == "" {
		return nil, errors.New("gmail: invalid notification addresses")
	}

	return Notifier(Gmail{
		name: config.Name,
		user: config.Config["user"],
		pass: config.Config["pass"],
		notificationAddresses: strings.Split(config.Config["notification_addresses"], ","),
	}), nil
}

func (a Gmail) Name() string {
	return a.name
}

func (a Gmail) Notify(msg Message) error {

	body := "To: " + strings.Join(a.notificationAddresses, ",") +
		"\r\nSubject: " + msg.ShortMessage() +
		"\r\n\r\n" + msg.ShortMessage()

	auth := smtp.PlainAuth("", a.user, a.pass, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, a.user,
		a.notificationAddresses, []byte(body))
	if err != nil {
		return err
	}

	return nil
}
