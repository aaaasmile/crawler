package db

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"testing"
)

func FilloutCost(t *testing.T) { // Prefix with test if you want to run this function again
	// function used to update the db from a csv file
	// Quantity and cost
	lite := LiteDB{
		SqliteDBPath: "../chart-info.db",
	}
	if err := lite.OpenSqliteDatabase(); err != nil {
		t.Error("Error on open db", err)
		return
	}

	fname := "../data/Depotansicht_EASYBANK_20068139598_20210124.csv"
	csvFile, err := os.Open(fname)
	if err != nil {
		t.Error("Open file error ", err)
		return
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.Comma = ';'
	fieldKeys := []string{"time", "ISIN", "Type", "Name", "Place", "Quantity", "QtyLBL", "ValAt", "ValCurr", "ValDate", "Cost", "CostCurr", "W/L", "W/LCurr", "W/LPerc", "W/LPercSymb"}

	lineCount := 0
	stocks := make([]*StockInfo, 0)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Error("Error on read", err)
			return
		}
		mmfield := make(map[string]string)
		for i, item := range line {
			k := fieldKeys[i]
			mmfield[k] = item
		}
		item := StockInfo{
			ISIN:     mmfield["ISIN"],
			Quantity: getRealFromString(mmfield["Quantity"]),
			Cost:     getRealFromString(mmfield["Cost"]),
		}
		stocks = append(stocks, &item)
		lineCount++
	}
	for _, v := range stocks {
		fmt.Println(v)
		if _, err := lite.updateStockinfo(v.ISIN, v.Quantity, v.Cost); err != nil {
			t.Error("Update error ", err)
			return
		}
	}
	t.Error("Finished") // force output
}
