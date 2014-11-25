package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type Twilio struct {
	accountSid   string
	authToken    string
	phoneNumbers []string
	twilioNumber string
}

func (a Twilio) Trigger(event *Event) (err error) {

	msg := event.ShortMessage()
	for _, num := range a.phoneNumbers {
		err = SendSMS(a.accountSid, a.authToken, num, a.twilioNumber, msg)
		if err != nil {
			return
		}
	}
	event.server.log.Println(white, "Twilio alert successfully triggered.", reset)
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
