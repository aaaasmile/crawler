package crawler

import "testing"

func TestParsePrice(t *testing.T) {
	str := "IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)"
	pi, err := parseForPriceInfo(str)
	if err != nil {
		t.Errorf("Price info not recognized on %q with error:\n %v", str, err)
		return
	}
	if pi.Price != 16.34 {
		t.Error("Expect 16.34 but recived ", pi.Price)
	}
	t.Error(pi.Timestamp)
}
