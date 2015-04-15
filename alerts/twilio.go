package alerts

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/jonog/redalert/core"
)

type TwilioConfig struct {
	AccountSID          string   `json:"account_sid"`
	AuthToken           string   `json:"auth_token"`
	TwilioNumber        string   `json:"twilio_number"`
	NotificationNumbers []string `json:"notification_numbers"`
}

type Twilio struct {
	accountSid   string
	authToken    string
	phoneNumbers []string
	twilioNumber string
}

func NewTwilio(config *TwilioConfig) Twilio {
	return Twilio{
		accountSid:   config.AccountSID,
		authToken:    config.AuthToken,
		phoneNumbers: config.NotificationNumbers,
		twilioNumber: config.TwilioNumber,
	}
}

func (a Twilio) Name() string {
	return "Twilio"
}

func (a Twilio) Trigger(event *core.Event) (err error) {

	msg := event.ShortMessage()
	for _, num := range a.phoneNumbers {
		err = SendSMS(a.accountSid, a.authToken, num, a.twilioNumber, msg)
		if err != nil {
			return
		}
	}
	event.Server.Log.Println("Twilio alert successfully triggered.")
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
