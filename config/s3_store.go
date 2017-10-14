package config

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

type S3Store struct {
	URL  string
	data S3StoreData
}

type S3StoreData struct {
	Checks        []checks.Config    `json:"checks"`
	Notifications []notifiers.Config `json:"notifications"`
	Preferences   Preferences        `json:"preferences"`
}

func NewS3Store(URL string) (*S3Store, error) {
	config := &S3Store{URL: URL}
	err := config.read()
	if err != nil {
		return nil, err
	}

	// create check ID if not present
	for i := range config.data.Checks {
		if config.data.Checks[i].ID == "" {
			config.data.Checks[i].ID = generateID(8)
		}
	}

	// create notification ID if not present
	for i := range config.data.Notifications {
		if config.data.Notifications[i].ID == "" {
			config.data.Notifications[i].ID = generateID(8)
		}
	}

	return config, nil
}

func (s *S3Store) read() error {

	u, err := url.Parse(s.URL)
	if err != nil {
		return err
	}
	log.Printf("bucket: %q, key: %q", u.Host, u.Path)

	sess := session.Must(session.NewSession())
	body, err := getS3File(sess, u.Host, u.Path)
	if err != nil {
		return err
	}

	var data S3StoreData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	s.data = data
	return nil
}

func getS3File(s *session.Session, bucket, key string) (value []byte, err error) {
	results, err := s3.New(s).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *S3Store) Notifications() ([]notifiers.Config, error) {
	return s.data.Notifications, nil
}

func (s *S3Store) Checks() ([]checks.Config, error) {
	return s.data.Checks, nil
}

func (s *S3Store) Preferences() (Preferences, error) {
	return s.data.Preferences, nil
}
