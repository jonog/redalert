package notifiers

import (
	"errors"
	"net/smtp"
	"strings"
)

type Gmail struct {
	name                  string
	user                  string
	pass                  string
	notificationAddresses []string
}

func NewGmailNotifier(config Config) (Gmail, error) {

	if config.Type != "gmail" {
		return Gmail{}, errors.New("gmail: invalid config type")
	}

	if config.Config["user"] == "" {
		return Gmail{}, errors.New("gmail: invalid user")
	}

	if config.Config["pass"] == "" {
		return Gmail{}, errors.New("gmail: invalid pass")
	}

	if config.Config["notification_addresses"] == "" {
		return Gmail{}, errors.New("gmail: invalid notification addresses")
	}

	return Gmail{
		name: config.Name,
		user: config.Config["user"],
		pass: config.Config["pass"],
		notificationAddresses: strings.Split(config.Config["notification_addresses"], ","),
	}, nil
}

func (a Gmail) Name() string {
	return a.name
}

func (a Gmail) Trigger(msg Message) error {

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
