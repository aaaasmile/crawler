package mail

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	mathrand "math/rand"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aaaasmile/crawler/conf"
	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/idl"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type MailSender struct {
	liteDB         *db.LiteDB
	gmailService   *gmail.Service
	secret         *db.Secret
	serviceAccount *conf.ServiceAccount
	simulate       bool
}

func NewMailSender(ld *db.LiteDB, sa *conf.ServiceAccount, simulate bool) *MailSender {
	ms := MailSender{
		liteDB:         ld,
		simulate:       simulate,
		serviceAccount: sa,
	}
	return &ms
}

func (ms *MailSender) FetchSecretFromDb() error {
	secr, err := ms.liteDB.FetchSecret()
	if err != nil {
		return err
	}
	log.Println("Secrets: ", secr)
	if len(secr) != 1 {
		return fmt.Errorf("Secret is not inserted or is multiple. Please check the db")
	}

	ms.secret = &secr[0]
	return nil
}

func (ms *MailSender) AuthGmailServiceWithDBSecret() error {
	accessToken := ms.secret.AccessToken
	if accessToken == "" {
		accessToken = ms.secret.AuthToken
	}

	if err := ms.oAuthGmailService(accessToken, ms.secret.RefreshToken); err != nil {
		return err
	}
	return nil
}

func (ms *MailSender) oAuthGmailService(accessToken, refreshToken string) error {
	log.Println("Authorize with oauth")
	config := oauth2.Config{
		ClientID:     ms.secret.ClientID,
		ClientSecret: ms.secret.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	log.Println("Using access token: ", accessToken)

	token := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = config.TokenSource(context.Background(), &token)
	var tokenUpdated *oauth2.Token
	tokenUpdated, err := tokenSource.Token()
	if err != nil {
		return err
	}
	if ms.secret.AccessToken != tokenUpdated.AccessToken {
		log.Println("Access token updated")
		log.Println("Update secret in db")
		if _, err := ms.liteDB.UpdateSecret(ms.secret.ID, tokenUpdated.AccessToken, tokenUpdated.RefreshToken); err != nil {
			return err
		}
	}

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return err
	}

	ms.gmailService = srv
	log.Println("Email service is initialized")
	return nil
}

func (ms *MailSender) SendEmailViaOAUTH2(templFileName string, listsrc []*idl.ChartInfo) error {
	log.Println("Send e-mail with gmail service using multipart. Charts: ", len(listsrc))
	if ms.gmailService == nil {
		return fmt.Errorf("Gmail service was not authorized and created")
	}
	bound1 := randomBoundary()
	bound2 := randomBoundary()

	list := make([]*idl.ChartInfo, 0)
	listErr := make([]*idl.ChartInfo, 0)
	imgBuf := &bytes.Buffer{}
	for _, v := range listsrc {
		if v.DownloadFilename == "" || v.HasError || v.ErrorText != "" {
			log.Println("Wrong img: ", v)
			listErr = append(list, v)
			continue
		}
		fname, err := embedImgFile(v.DownloadFilename, imgBuf, bound1)
		if err != nil {
			log.Println("Ignore image ", v, err)
			v.ErrorText = err.Error()
			listErr = append(list, v)
		} else {
			v.ImgName = fname
			list = append(list, v)
		}
	}
	if len(list) > 0 {
		imgBuf.Write([]byte("--" + bound1 + "--"))
	}
	if len(listErr) == 0 {
		log.Println("wow, all images are ok ", len(list))
	} else {
		log.Printf("Some errors: ok %d, error %d\n", len(list), len(listErr))
	}

	ctx := struct {
		ListOK  []*idl.ChartInfo
		ListErr []*idl.ChartInfo
	}{
		ListOK:  list,
		ListErr: listErr,
	}

	var partHTMLCont, partSubj, partPlainContent bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partHTMLCont, "mailbody", ctx); err != nil {
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

	// plain section
	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/plain; charset=\"UTF-8\"\r\n"))
	partPlainContent.WriteTo(msg)
	msg.Write([]byte("\r\n"))

	// html section
	msg.Write([]byte("--" + bound2 + "\r\n"))
	msg.Write([]byte("Content-Type: text/html; charset=\"UTF-8\"\r\n"))
	partHTMLCont.WriteTo(msg)
	msg.Write([]byte("\r\n"))
	msg.Write([]byte("--" + bound2 + "--" + "\r\n"))

	// embedded images section
	imgBuf.WriteTo(msg)

	if ms.simulate {
		ss := msg.String()
		maxchar := 1200
		if len(ss) > maxchar {
			ss = ss[0:maxchar]
		}
		fmt.Printf("Message is: \n%s\n", ss)
	}

	message.Raw = base64.URLEncoding.EncodeToString(msg.Bytes())

	if !ms.simulate {
		if _, err := ms.gmailService.Users.Messages.Send("me", &message).Do(); err != nil {
			return err
		}

		log.Println("E-Mail is on the way. Everything is going well.")
	} else {
		log.Printf("Simulate Mail sent")
	}

	return nil
}

func embedImgFile(fullname string, w *bytes.Buffer, boundary string) (string, error) {
	log.Println("Processing ", fullname)
	if _, err := os.Stat(fullname); err != nil {
		log.Println("File error on ", fullname, err)
		return "", err
	}
	raw, err := ioutil.ReadFile(fullname)
	if err != nil {
		log.Println("Read file error: ", fullname, err)
		return "", err
	}
	fname := filepath.Base(fullname)
	extimg := strings.ToLower(filepath.Ext(fname))
	if !strings.HasSuffix(extimg, "jpg") && !strings.HasSuffix(extimg, "png") {
		log.Println("Image not supported ", extimg)
		return "", err
	}

	xname := fmt.Sprintf("ii_kj%s", randomIdAscii(8))
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

	return xname, nil

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

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func randomIdAscii(size int) string {
	set := make([]int, 0)
	for i := 48; i < 58; i++ {
		set = append(set, i)
	}
	for i := 97; i < 123; i++ {
		set = append(set, i)
	}
	buf := make([]byte, 0)
	for i := 0; i < size; i++ {
		ixrnd := mathrand.Intn(len(set))
		buf = append(buf, byte(set[ixrnd]))
	}
	return string(buf)
}
