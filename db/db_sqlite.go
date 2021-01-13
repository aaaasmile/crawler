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

func (ld *LiteDB) FetchSecret(pageIx int, pageSize int) ([]Secret, error) {
	q := `SELECT id,clientid,clientsecret,authtoken,refreshtoken
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

	defer rows.Close()
	res := make([]Secret, 0)
	for rows.Next() {
		item := Secret{}
		if err := rows.Scan(&item.ID, &item.ClientID, &item.ClientSecret, &item.AuthToken, &item.RefreshToken); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
