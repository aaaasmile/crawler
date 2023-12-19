package crawler

import "testing"

func TestParsePrice(t *testing.T) {
	//str := "IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)"
	pricestr := "16,34"
	closed := "Schluss 15.01.21"
	pi, err := parseForPriceInfo(pricestr, closed)
	if err != nil {
		t.Errorf("Price info not recognized on %q %q with error:\n %v", pricestr, closed, err)
		return
	}
	if pi.Price != 16.34 {
		t.Error("Expect 16.34 but recived ", pi.Price)
	}
	if pi.TimestampInt != 1610732160 {
		t.Error("Expect 1610732160 but recived ", pi.TimestampInt)
	}

}
