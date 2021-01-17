package idl

var (
	Appname = "Chart Crawler"
	Buildnr = "0.1.20210114-00"
)

type ChartInfo struct {
	Fname       string
	Fullname    string
	Description string
	HasError    bool
	ErrorText   string
	Alt         string
	MoreInfoURL string
}
