package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	RelayMail    string
	RelaySecret  string
	RelayHost    string
	RelayUser    string
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
	Quantity    float64
	Cost        float64
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

func getRealFromString(s string) float64 {
	// s is something like 2.639,00
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	res, _ := strconv.ParseFloat(s, 64)
	return res
}

func (sc *Secret) FromNullString(ci, cs, aut, rt, em, at, relaymail, realaysecret, relayhost, relayuser sql.NullString) {
	sc.ClientID = nullStrToStr(ci)
	sc.ClientSecret = nullStrToStr(cs)
	sc.AuthToken = nullStrToStr(aut)
	sc.RefreshToken = nullStrToStr(rt)
	sc.Email = nullStrToStr(em)
	sc.AccessToken = nullStrToStr(at)
	sc.RelayHost = nullStrToStr(relayhost)
	sc.RelayMail = nullStrToStr(relaymail)
	sc.RelaySecret = nullStrToStr(realaysecret)
	sc.RelayUser = nullStrToStr(relayuser)
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
	q := `SELECT id,clientid,clientsecret,authtoken,refreshtoken,email,accesstoken,relaymail,realaysecret,relayhost,relayuser
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
	var relaymail, realaysecret, relayhost, relayuser sql.NullString
	defer rows.Close()
	res := make([]Secret, 0)
	for rows.Next() {
		item := Secret{}
		if err := rows.Scan(&item.ID, &ci, &cs,
			&aut, &rt, &em, &at, &relaymail, &realaysecret, &relayhost, &relayuser); err != nil {
			return nil, err
		}
		item.FromNullString(ci, cs, aut, rt, em, at, relaymail, realaysecret, relayhost, relayuser)
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
	q := `SELECT id,isin,charturl,name,description,moreinfourl,quantity,cost
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
			&na, &des, &mor, &item.Quantity, &item.Cost); err != nil {
			return nil, err
		}
		item.FromNullString(isin, cu, na, des, mor)
		res = append(res, &item)
	}
	return res, nil
}

func (ld *LiteDB) updateStockinfo(isin string, quantity, cost float64) (int64, error) {
	q := `UPDATE stockinfo SET (quantity,cost) = (?,?)
	WHERE isin='%s';`
	q = fmt.Sprintf(q, isin)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	stmt, err := ld.connDb.Prepare(q)
	if err != nil {
		return 0, err
	}

	ressql, err := stmt.Exec(quantity, cost)
	if err != nil {
		return 0, err
	}
	rowNr, _ := ressql.RowsAffected()
	log.Println("update, rows affected: ", rowNr)

	return rowNr, nil
}

func (ld *LiteDB) InsertPrice(tx *sql.Tx, idstock int64, price float64, timestamp int64) (int64, error) {
	q := fmt.Sprintf(`INSERT INTO price(idstock,price,timestamp) VALUES(?,?,?)`)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}

	stmt, err := ld.connDb.Prepare(q)
	if err != nil {
		return 0, err
	}
	ressql, err := tx.Stmt(stmt).Exec(idstock, price, timestamp)
	if err != nil {
		return 0, err
	}

	return ressql.LastInsertId()
}

func (ld *LiteDB) FetchPrice(idstock int64, price float64, timestamp int64) ([]*Price, error) {
	q := `SELECT id,price,timestamp,idstock
		  FROM price
		  WHERE idstock=%d AND price=%f AND timestamp=%d
		  LIMIT 1;`
	q = fmt.Sprintf(q, idstock, price, timestamp)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}
	return ld.selectQueryPrice(q)
}

func (ld *LiteDB) FetchPreviosPriceInStock(idstock int64, timestampLatest int64) ([]*Price, error) {
	q := `SELECT id,price,timestamp,idstock
		  FROM price
		  WHERE idstock=%d AND timestamp<%d
		  LIMIT 1;`
	q = fmt.Sprintf(q, idstock, timestampLatest)
	if ld.DebugSQL {
		log.Println("Query is", q)
	}
	return ld.selectQueryPrice(q)
}

func (ld *LiteDB) selectQueryPrice(q string) ([]*Price, error) {
	rows, err := ld.connDb.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]*Price, 0)
	for rows.Next() {
		item := Price{}
		var ti int64
		if err := rows.Scan(&item.ID, &item.Price, &ti, &item.IDStock); err != nil {
			return nil, err
		}
		item.Timestamp = time.Unix(ti, 0)
		res = append(res, &item)
	}
	return res, nil
}

func (ld *LiteDB) GetNewTransaction() (*sql.Tx, error) {
	tx, err := ld.connDb.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (ld *LiteDB) CommitTransaction(tx *sql.Tx) error {
	return tx.Commit()
}
