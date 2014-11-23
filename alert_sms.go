package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type SMS struct {
	accountSid   string
	authToken    string
	phoneNumber  string
	twilioNumber string
}

func (a SMS) Send(server *Server) error {

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + a.accountSid + "/Messages.json"

	v := url.Values{}
	v.Set("To", a.phoneNumber)
	v.Set("From", a.twilioNumber)
	v.Set("Body", "Uhoh, "+server.name+" has been nuked!!!")
	rb := *strings.NewReader(v.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &rb)
	req.SetBasicAuth(a.accountSid, a.authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return errors.New("Invalid Twilio status code")
	}
	return err

}
