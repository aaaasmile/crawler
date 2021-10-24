package idl

import "github.com/aaaasmile/crawler/db"

var (
	Appname = "Chart Crawler"
	Buildnr = "0.1.20210114-00"
)

type ChartInfo struct {
	DownloadFilename  string
	ImgName           string
	CurrentPrice      string
	PriceInfo         *db.Price
	PreviousPrice     float64
	DiffPreviousPrice float64
	WinOrLoss         string
	WinOrLossPerc     string
	HasError          bool
	ErrorText         string
	Description       string
	MoreInfoURL       string
	ChartURL          string
	ID                int64
	TotCurrValue      string
	TotCost           string
	Quantity          string
	ColorWL           string
	SimpleDescr       string
}
