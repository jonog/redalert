package notifiers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	registerNotifier("twilio", NewTwilioNotifier)
}

type Twilio struct {
	accountSid   string
	authToken    string
	phoneNumbers []string
	twilioNumber string
}

var NewTwilioNotifier = func(config Config) (Notifier, error) {

	if config.Type != "twilio" {
		return nil, errors.New("twilio: invalid config type")
	}

	if config.Config["account_sid"] == "" {
		return nil, errors.New("twilio: invalid account_sid")
	}

	if config.Config["auth_token"] == "" {
		return nil, errors.New("twilio: invalid auth_token")
	}

	if config.Config["twilio_number"] == "" {
		return nil, errors.New("twilio: invalid twilio_number")
	}

	if config.Config["notification_numbers"] == "" {
		return nil, errors.New("twilio: invalid notification_numbers")
	}

	return Notifier(Twilio{
		accountSid:   config.Config["account_sid"],
		authToken:    config.Config["auth_token"],
		phoneNumbers: strings.Split(config.Config["notification_numbers"], ","),
		twilioNumber: config.Config["twilio_number"],
	}), nil
}

func (a Twilio) Name() string {
	return "Twilio"
}

func (a Twilio) Notify(msg Message) (err error) {

	smsText := msg.DefaultMessage
	for _, num := range a.phoneNumbers {
		err = SendSMS(a.accountSid, a.authToken, num, a.twilioNumber, smsText)
		if err != nil {
			return
		}
	}

	return nil

}

func SendSMS(accountSID string, authToken string, to string, from string, body string) error {

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSID + "/Messages.json"

	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Body", body)
	rb := *strings.NewReader(v.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &rb)
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return errors.New("Invalid Twilio status code")
	}
	return err

}
