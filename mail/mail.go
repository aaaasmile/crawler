package mail

import (
	"log"

	"github.com/aaaasmile/crawler/db"
)

type MailSender struct {
	liteDB *db.LiteDB
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

	return nil
}
