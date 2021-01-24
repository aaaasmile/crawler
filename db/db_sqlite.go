package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
	AccessToken  string
	RefreshToken string
	Email        string
}

type Price struct {
	ID           int64
	Price        float64
	TimestampInt int64
	Timestamp    time.Time
	IDStock      int64
}

type StockInfo struct {
	ID          int64
	ISIN        string
	ChartURL    string
	Name        string
	Description string
	MoreInfoURL string
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

func (sc *Secret) FromNullString(ci, cs, aut, rt, em, at sql.NullString) {
	sc.ClientID = nullStrToStr(ci)
	sc.ClientSecret = nullStrToStr(cs)
	sc.AuthToken = nullStrToStr(aut)
	sc.RefreshToken = nullStrToStr(rt)
	sc.Email = nullStrToStr(em)
	sc.AccessToken = nullStrToStr(at)
}

func (si *StockInfo) FromNullString(isin, cu, na, des, mor sql.NullString) {
	si.ISIN = nullStrToStr(isin)
	si.ChartURL = nullStrToStr(cu)
	si.Name = nullStrToStr(na)
	si.Description = nullStrToStr(des)
	si.MoreInfoURL = nullStrToStr(mor)
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
	q := `SELECT id,clientid,clientsecret,authtoken,refreshtoken,email,accesstoken
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

	var ci, cs, aut, rt, em, at sql.NullString

	defer rows.Close()
	res := make([]Secret, 0)
	for rows.Next() {
		item := Secret{}
		if err := rows.Scan(&item.ID, &ci, &cs,
			&aut, &rt, &em, &at); err != nil {
			return nil, err
		}
		item.FromNullString(ci, cs, aut, rt, em, at)
		res = append(res, item)
	}
	return res, nil
}

func (ld *LiteDB) UpdateSecret(ID int, accessToken, refreshToken string) (int64, error) {
	q := `UPDATE secrets SET (accesstoken,refreshtoken) = (?,?)
	WHERE id=%d;`
	q = fmt.Sprintf(q, ID)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	stmt, err := ld.connDb.Prepare(q)
	if err != nil {
		return 0, err
	}

	ressql, err := stmt.Exec(accessToken, refreshToken)
	if err != nil {
		return 0, err
	}
	rowNr, _ := ressql.RowsAffected()
	log.Println("update, rows affected: ", rowNr)

	return rowNr, nil

}

func (ld *LiteDB) FetchStockInfo(limit int) ([]*StockInfo, error) {
	q := `SELECT id,isin,charturl,name,description,moreinfourl
		  FROM stockinfo
		  LIMIT %d;`
	q = fmt.Sprintf(q, limit)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	rows, err := ld.connDb.Query(q)
	if err != nil {
		return nil, err
	}

	var isin, cu, na, des, mor sql.NullString

	defer rows.Close()
	res := make([]*StockInfo, 0)
	for rows.Next() {
		item := StockInfo{}
		if err := rows.Scan(&item.ID, &isin, &cu,
			&na, &des, &mor); err != nil {
			return nil, err
		}
		item.FromNullString(isin, cu, na, des, mor)
		res = append(res, &item)
	}
	return res, nil
}

func (ld *LiteDB) InsertPrice(idstock int64, price float64, timestamp int64) (int64, error) {
	q := fmt.Sprintf(`INSERT INTO price(idstock,price,timestamp) VALUES(?,?,?)`)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	stmt, err := ld.connDb.Prepare(q)
	if err != nil {
		return 0, err
	}
	ressql, err := stmt.Exec(idstock, price, timestamp)
	if err != nil {
		return 0, err
	}

	return ressql.LastInsertId()
}
