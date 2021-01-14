package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/idl"
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

func (ms *MailSender) SendEmailOAUTH2(templFileName string, listsrc []*idl.ChartInfo) error {
	log.Println("Send e-mail with gmail service")

	list := make([]*idl.ChartInfo, 0, len(listsrc))
	imgBuf := &bytes.Buffer{}
	for _, v := range listsrc {
		if fname := embedImgFile(v.Fullname, imgBuf); fname != "" {
			v.Fname = fname
			list = append(list, v)
		} else {
			log.Println("Ignore image ", v)
		}
	}

	var partContent, partSubj bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partContent, "mailbody", list); err != nil {
		return err
	}
	if err := tmplBodyMail.ExecuteTemplate(&partSubj, "mailSubj", list); err != nil {
		return err
	}

	var message gmail.Message

	msg := &bytes.Buffer{}
	emailTo := []byte("To: " + ms.secret.Email + "\r\n")
	mime := []byte("MIME-version: 1.0;\n")
	msg.Write(emailTo)
	msg.Write(partSubj.Bytes())
	msg.Write(mime)
	imgBuf.WriteTo(msg)
	msg.Write([]byte("Content-Type: text/html; charset=\"UTF-8\";\n\n"))
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

func embedImgFile(fullname string, w *bytes.Buffer) string {
	log.Println("Processing ", fullname)
	if _, err := os.Stat(fullname); err != nil {
		log.Println("File error on ", fullname, err)
		return ""
	}
	raw, err := ioutil.ReadFile(fullname)
	if err != nil {
		log.Println("Read file error: ", fullname, err)
		return ""
	}

	fname := filepath.Base(fullname)
	mediaType := mime.TypeByExtension(filepath.Ext(fname))
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	w.Write([]byte("Content-Type: " + mediaType + `; name="` + fname + `"` + "\r\n"))
	w.Write([]byte("Content-Transfer-Encoding: Base64 \r\n"))
	w.Write([]byte("Content-Disposition: inline" + `; name="` + fname + `"` + "\r\n"))
	w.Write([]byte("Content-ID:" + "<" + fname + ">" + "\r\n"))
	w.Write(Base64Encode(raw))
	w.Write([]byte("\r\n"))

	return fname
	// if _, ok := f.Header["Content-Transfer-Encoding"]; !ok {
	// 		f.setHeader("Content-Transfer-Encoding", string(Base64))
	// 	}

	// 	if _, ok := f.Header["Content-Disposition"]; !ok {
	// 		var disp string
	// 		disp = "inline"
	// 		f.setHeader("Content-Disposition", disp+`; filename="`+f.Name+`"`)
	// 	}

	// 	if !isAttachment {
	// 		if _, ok := f.Header["Content-ID"]; !ok {
	// 			f.setHeader("Content-ID", "<"+f.Name+">")
	// 		}
	// 	}
	// 	w.writeHeaders(f.Header)
	// 	w.writeBody(f.CopyFunc, Base64)
	// }
}

func Base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}
