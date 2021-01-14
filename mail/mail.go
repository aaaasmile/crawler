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
	GmailService *gmail.Service
	Secret       *db.Secret
}

func NewMailSender(ld *db.LiteDB) (*MailSender, error) {
	ms := MailSender{
		liteDB: ld,
	}
	secr, err := ms.liteDB.FetchSecret()
	if err != nil {
		return nil, err
	}
	log.Println("Secrets: ", secr)
	if len(secr) != 1 {
		return nil, fmt.Errorf("Secret is not inserted or is multiple. Please check the db")
	}
	ms.Secret = &secr[0]
	ms.oAuthGmailService()

	return &ms, nil
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
	emailTo := []byte("To: " + ms.Secret.Email + "\r\n")
	mime := []byte("MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n")
	msg.Write(emailTo)
	msg.Write(partSubj.Bytes())
	msg.Write(mime)
	msg.Write(partContent.Bytes())

	fmt.Println("*** Message is: ", msg.String())

	message.Raw = base64.URLEncoding.EncodeToString(msg.Bytes())

	// if _, err := ms.GmailService.Users.Messages.Send("me", &message).Do(); err != nil {
	// 	return err
	// }

	log.Println("E-Mail is on the way. Everything is going well.")
	return nil
}
