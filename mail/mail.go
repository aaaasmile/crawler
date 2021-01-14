package mail

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aaaasmile/crawler/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type MailSender struct {
	liteDB       *db.LiteDB
	GmailService *gmail.Service
	Secret       *db.Secret
}

func NewMailSender(ld *db.LiteDB) *MailSender {
	res := MailSender{
		liteDB: ld,
	}
	return &res
}

func (ms *MailSender) SendEmail() error {
	log.Println("Send email")
	secr, err := ms.liteDB.FetchSecret()
	if err != nil {
		return err
	}
	log.Println("Secrets: ", secr)
	if len(secr) != 1 {
		return fmt.Errorf("Secret is not inserted or is multiple. PLease check the db")
	}
	ms.Secret = &secr[0]
	ms.oAuthGmailService()
	ms.sendEmailOAUTH2()

	return nil
}

func (ms *MailSender) oAuthGmailService() {
	log.Println("Authorize with oauth")
	config := oauth2.Config{
		ClientID:     ms.Secret.ClientID,
		ClientSecret: ms.Secret.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	token := oauth2.Token{
		AccessToken:  ms.Secret.AuthToken,
		RefreshToken: ms.Secret.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = config.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
	}

	ms.GmailService = srv
	if ms.GmailService != nil {
		fmt.Println("Email service is initialized \n")
	}
}

func (ms *MailSender) sendEmailOAUTH2() (bool, error) {
	log.Println("Send e-mail with gmail service")
	var err error
	// emailBody, err := parseTemplate(template, data)
	// if err != nil {
	// 	return false, errors.New("unable to parse email template")
	// }
	emailBody := "Questa è una mail di charts"

	var message gmail.Message

	emailTo := "To: " + ms.Secret.Email + "\r\n"
	subject := "Subject: " + "Test Email form Gmail API using OAuth" + "\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(emailTo + subject + mime + "\n" + emailBody)

	message.Raw = base64.URLEncoding.EncodeToString(msg)

	// Send the message
	_, err = ms.GmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return false, err
	}
	log.Println("E-Mail is on the way")
	return true, nil
}
