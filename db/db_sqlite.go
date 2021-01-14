package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type LiteDB struct {
	connDb       *sql.DB
	DebugSQL     bool
	SqliteDBPath string
}

type Secret struct {
	ID           int
	ClientID     string
	ClientSecret string
	AuthToken    string
	RefreshToken string
	Email        string
}

func nullStrToStr(ci sql.NullString) string {
	if ci.Valid {
		v, _ := ci.Value()
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	return ""
}

func (sc *Secret) FromNullString(ci, cs, aut, rt, em sql.NullString) {
	sc.ClientID = nullStrToStr(ci)
	sc.ClientSecret = nullStrToStr(cs)
	sc.AuthToken = nullStrToStr(aut)
	sc.RefreshToken = nullStrToStr(rt)
	sc.Email = nullStrToStr(em)
}

func (ld *LiteDB) OpenSqliteDatabase() error {
	var err error
	dbname := ld.SqliteDBPath
	if _, err := os.Stat(dbname); err != nil {
		return err
	}
	log.Println("Using the sqlite file: ", dbname)
	ld.connDb, err = sql.Open("sqlite3", dbname)
	if err != nil {
		return err
	}
	return nil
}

func (ld *LiteDB) FetchSecret() ([]Secret, error) {
	q := `SELECT id,clientid,clientsecret,authtoken,refreshtoken,email
		  FROM secrets
		  LIMIT 1;`
	q = fmt.Sprintf(q)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	rows, err := ld.connDb.Query(q)
	if err != nil {
		return nil, err
	}

	var ci, cs, aut, rt, em sql.NullString

	defer rows.Close()
	res := make([]Secret, 0)
	for rows.Next() {
		item := Secret{}
		if err := rows.Scan(&item.ID, &ci, &cs,
			&aut, &rt, &em); err != nil {
			return nil, err
		}
		item.FromNullString(ci, cs, aut, rt, em)
		res = append(res, item)
	}
	return res, nil
}
