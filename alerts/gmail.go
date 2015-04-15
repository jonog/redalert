package alerts

import (
	"net/smtp"
	"strings"

	"github.com/jonog/redalert/core"
)

type GmailConfig struct {
	User                  string   `json:"user"`
	Pass                  string   `json:"pass"`
	NotificationAddresses []string `json:"notification_addresses"`
}

type Gmail struct {
	user                  string
	pass                  string
	notificationAddresses []string
}

func NewGmail(config *GmailConfig) Gmail {
	return Gmail{
		user: config.User,
		pass: config.Pass,
		notificationAddresses: config.NotificationAddresses,
	}
}

func (a Gmail) Name() string {
	return "Gmail"
}

func (a Gmail) Trigger(event *core.Event) error {

	body := "To: " + strings.Join(a.notificationAddresses, ",") +
		"\r\nSubject: " + event.ShortMessage() +
		"\r\n\r\n" + event.ShortMessage()

	auth := smtp.PlainAuth("", a.user, a.pass, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, a.user,
		a.notificationAddresses, []byte(body))
	if err != nil {
		return err
	}

	event.Server.Log.Println("Gmail alert successfully triggered.")
	return nil
}
