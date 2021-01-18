package idl

var (
	Appname = "Chart Crawler"
	Buildnr = "0.1.20210114-00"
)

type ChartInfo struct {
	DownloadFilename string
	ImgName          string
	CurrentPrice     string
	HasError         bool
	ErrorText        string
	Description      string
	MoreInfoURL      string
}
