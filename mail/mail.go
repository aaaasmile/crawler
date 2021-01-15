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
	"strings"
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

	bound1 := "000000000000d22e8805b8e517cc" // TODO
	bound2 := "000000000000d22e8605b8e517cb" // TODO

	list := make([]*idl.ChartInfo, 0, len(listsrc))
	imgBuf := &bytes.Buffer{}
	for _, v := range listsrc {
		if fname := embedImgFile(v.Fullname, imgBuf, bound1); fname != "" {
			v.Fname = fname
			list = append(list, v)
		} else {
			log.Println("Ignore image ", v)
		}
	}

	var partHtmlContent, partSubj, partPlainContent bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partHtmlContent, "mailbody", list); err != nil {
		return err
	}
	if err := tmplBodyMail.ExecuteTemplate(&partSubj, "mailSubj", list); err != nil {
		return err
	}

	if err := tmplBodyMail.ExecuteTemplate(&partPlainContent, "mailPlain", list); err != nil {
		return err
	}

	var message gmail.Message

	msg := &bytes.Buffer{}
	msg.Write([]byte("MIME-version: 1.0;\r\n"))
	partSubj.WriteTo(msg)
	msg.Write([]byte("To: " + ms.secret.Email + "\r\n"))
	msg.Write([]byte("Content-Type:  multipart/related; boundary=" + `"` + bound1 + `"` + "\r\n"))
	msg.Write([]byte("\r\n"))

	msg.Write([]byte("--" + bound1 + "\r\n"))
	msg.Write([]byte("Content-Type:  multipart/alternative; boundary=" + `"` + bound2 + `"` + "\r\n"))
	msg.Write([]byte("\r\n"))

	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/plain; charset=\"UTF-8\"\r\n"))
	partPlainContent.WriteTo(msg)
	msg.Write([]byte("\r\n"))

	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/html; charset=\"UTF-8\"\r\n"))
	partHtmlContent.WriteTo(msg)
	msg.Write([]byte("\r\n"))
	msg.Write([]byte("--" + bound2 + "--" + "\r\n"))
	imgBuf.WriteTo(msg)

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

func embedImgFile(fullname string, w *bytes.Buffer, boundary string) string {
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
	extimg := strings.ToLower(filepath.Ext(fname))
	if !strings.HasSuffix(extimg, "jpg") && !strings.HasSuffix(extimg, "png") {
		log.Println("Image not supported ", extimg)
		return ""
	}

	xname := "ii_kjxipppu0" // todo calculate
	rawForm76 := formatRFCRawWithEnc64(raw)

	mediaType := mime.TypeByExtension(extimg)
	w.Write([]byte("--" + boundary + "\r\n"))
	w.Write([]byte("Content-Type: " + mediaType + `; name="` + fname + `"` + "\r\n"))
	w.Write([]byte("Content-Disposition: attachment" + `; filename="` + fname + `"` + "\r\n"))
	w.Write([]byte("Content-Transfer-Encoding: base64 \r\n"))
	w.Write([]byte("X-Attachment-Id: " + xname + "\r\n"))
	w.Write([]byte("Content-ID: <" + xname + ">" + "\r\n"))
	w.Write([]byte("\r\n"))
	rawForm76.WriteTo(w)
	w.Write([]byte("\r\n"))
	w.Write([]byte("--" + boundary + "--"))

	return xname

}

func formatRFCRawWithEnc64(raw []byte) *bytes.Buffer {
	//  RFC 2045 formatting to 76 col
	maxLineLen := 76
	p := base64Encode(raw)
	w := &bytes.Buffer{}
	n := 0
	lineLen := 0
	for len(p)+lineLen > maxLineLen {
		w.Write(p[:maxLineLen-lineLen])
		w.Write([]byte("\r\n"))
		p = p[maxLineLen-lineLen:]
		n += maxLineLen - lineLen
		lineLen = 0
	}
	w.Write(p)
	log.Println("Buffer size: ", n+len(p))

	return w
}

func base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}
