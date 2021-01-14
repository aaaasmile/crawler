package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
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
	gmailService *gmail.Service
	secret       *db.Secret
	simulate     bool
}

func NewMailSender(ld *db.LiteDB, simulate bool) (*MailSender, error) {
	ms := MailSender{
		liteDB:   ld,
		simulate: simulate,
	}
	secr, err := ms.liteDB.FetchSecret()
	if err != nil {
		return nil, err
	}
	log.Println("Secrets: ", secr)
	if len(secr) != 1 {
		return nil, fmt.Errorf("Secret is not inserted or is multiple. Please check the db")
	}
	ms.secret = &secr[0]
	ms.oAuthGmailService()

	return &ms, nil
}

func (ms *MailSender) oAuthGmailService() {
	log.Println("Authorize with oauth")
	config := oauth2.Config{
		ClientID:     ms.secret.ClientID,
		ClientSecret: ms.secret.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	token := oauth2.Token{
		AccessToken:  ms.secret.AuthToken,
		RefreshToken: ms.secret.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = config.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
	}

	ms.gmailService = srv
	if ms.gmailService != nil {
		fmt.Println("Email service is initialized \n")
	}
}

func (ms *MailSender) SendEmailOAUTH2(templFileName string, ctx interface{}) error {
	log.Println("Send e-mail with gmail service")

	var partContent, partSubj bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partContent, "mailbody", ctx); err != nil {
		return err
	}
	if err := tmplBodyMail.ExecuteTemplate(&partSubj, "mailSubj", ctx); err != nil {
		return err
	}

	var message gmail.Message
	msg := &bytes.Buffer{}
	emailTo := []byte("To: " + ms.secret.Email + "\r\n")
	mime := []byte("MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n")
	msg.Write(emailTo)
	msg.Write(partSubj.Bytes())
	msg.Write(mime)
	msg.Write(partContent.Bytes())

	fmt.Println("*** Message is: ", msg.String())

	message.Raw = base64.URLEncoding.EncodeToString(msg.Bytes())

	if !ms.simulate {
		if _, err := ms.gmailService.Users.Messages.Send("me", &message).Do(); err != nil {
			return err
		}
		log.Println("E-Mail is on the way. Everything is going well.")
	} else {
		log.Println("Simulate Mail sent")
	}

	return nil
}
