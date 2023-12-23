package mail

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	mathrand "math/rand"
	"mime"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaaasmile/crawler/db"
	"github.com/aaaasmile/crawler/idl"
)

type MailSender struct {
	liteDB    *db.LiteDB
	secret    *db.Secret
	simulate  bool
	emailTo   string
	emailFrom string
}

func NewMailSender(ld *db.LiteDB, simulate bool) *MailSender {
	ms := MailSender{
		liteDB:   ld,
		simulate: simulate,
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
		return fmt.Errorf("secret is not inserted or is multiple. Please check the db")
	}

	ms.secret = &secr[0]
	return nil
}

func (ms *MailSender) SendEmailViaRelay(templFileName string, listsrc []*idl.ChartInfo) error {
	log.Println("Send email using relay host")
	ms.emailTo = ms.secret.Email
	ms.emailFrom = ms.secret.RelayMail

	message, err := ms.buildEmailMsg(templFileName, listsrc)
	if err != nil {
		return err
	}
	if ms.simulate {
		log.Println("This is a simulation, e-mail is not sent")
		return nil
	}

	servername := ms.secret.RelayHost

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", ms.secret.RelayUser, ms.secret.RelaySecret, host)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	log.Println("Dial server ", servername)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	log.Println("Send smtp Auth")
	if err = c.Auth(auth); err != nil {
		return err
	}

	log.Println("send From")
	if err = c.Mail(ms.secret.RelayMail); err != nil {
		return err
	}
	log.Println("send To")
	if err = c.Rcpt(ms.secret.Email); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	log.Println("Send the message to the relay")
	_, err = w.Write(message.Bytes())
	if err != nil {
		return err
	}
	log.Println("Close relay")
	err = w.Close()
	if err != nil {
		return err
	}
	log.Println("Quit relay")
	c.Quit()
	log.Println("E-Mail is on the way. Everything is going well.")

	return nil
}

func (ms *MailSender) buildEmailMsg(templFileName string, listsrc []*idl.ChartInfo) (*bytes.Buffer, error) {
	bound1 := randomBoundary()
	bound2 := randomBoundary()

	list := make([]*idl.ChartInfo, 0)
	listObservation := make([]*idl.ChartInfo, 0)
	listErr := make([]*idl.ChartInfo, 0)
	imgBuf := &bytes.Buffer{}
	for _, v := range listsrc {
		if v.HasError || v.ErrorText != "" {
			log.Println("[WARN] Wrong img: ", v)
			listErr = append(listErr, v)
			continue
		}
		insert := false
		if v.DownloadFilename != "" {
			fname, err := embedImgFile(v.DownloadFilename, imgBuf, bound1)
			if err != nil {
				log.Println("Ignore image ", v, err)
				v.ErrorText = err.Error()
				listErr = append(listErr, v)
			} else {
				v.ImgName = fname
				insert = true
			}
		} else {
			log.Println("[WARN] image not found for ", v.ID, v.SimpleDescr)
			insert = true
		}
		if insert {
			if v.Quantity == "" || v.Quantity == "0.0" || v.Quantity == "0" || v.Quantity == "0.00" {
				listObservation = append(listObservation, v)
			} else {
				list = append(list, v)
			}
		}
	}
	if len(list) > 0 || len(listErr) > 0 {
		imgBuf.Write([]byte("--" + bound1 + "--"))
	}
	if len(listErr) == 0 {
		log.Println("wow, all images are ok ", len(list))
	} else {
		log.Printf("Some errors: ok %d, error %d\n", len(list), len(listErr))
	}

	ctx := struct {
		ListOK          []*idl.ChartInfo
		ListErr         []*idl.ChartInfo
		ListObservation []*idl.ChartInfo
	}{
		ListOK:          list,
		ListErr:         listErr,
		ListObservation: listObservation,
	}

	var partHTMLCont, partSubj, partPlainContent bytes.Buffer
	tmplBodyMail := template.Must(template.New("MailBody").ParseFiles(templFileName))
	if err := tmplBodyMail.ExecuteTemplate(&partHTMLCont, "mailbody", ctx); err != nil {
		return nil, err
	}
	if err := tmplBodyMail.ExecuteTemplate(&partSubj, "mailSubj", list); err != nil {
		return nil, err
	}

	if err := tmplBodyMail.ExecuteTemplate(&partPlainContent, "mailPlain", list); err != nil {
		return nil, err
	}

	msg := &bytes.Buffer{}
	msg.Write([]byte("MIME-version: 1.0;\r\n"))
	partSubj.WriteTo(msg)
	if ms.emailFrom != "" {
		msg.Write([]byte("From: " + ms.emailFrom + "\r\n"))
	}
	msg.Write([]byte("To: " + ms.emailTo + "\r\n"))
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
	msg.Write([]byte("Content-Transfer-Encoding: base64\r\n"))
	msg.Write([]byte("\r\n"))
	partHTMLCont64 := formatRFCRawWithEnc64(partHTMLCont.Bytes())
	partHTMLCont64.WriteTo(msg)
	msg.Write([]byte("\r\n"))
	msg.Write([]byte("--" + bound2 + "--" + "\r\n"))

	// embedded images section
	imgBuf.WriteTo(msg)

	if ms.simulate {
		ss := msg.String()
		maxchar := 2000
		if len(ss) > maxchar {
			ss = ss[0:maxchar]
		}
		fmt.Printf("Message is: \n%s\n", ss)
	}

	return msg, nil
}

func embedImgFile(fullname string, w *bytes.Buffer, boundary string) (string, error) {
	log.Println("Processing ", fullname)
	if _, err := os.Stat(fullname); err != nil {
		log.Println("File error on ", fullname, err)
		return "", err
	}
	raw, err := os.ReadFile(fullname)
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
