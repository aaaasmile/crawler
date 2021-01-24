package crawler

import "testing"

func TestParsePrice(t *testing.T) {
	str := "IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)"
	pi := parseForPriceInfo(str)
	if pi == nil {
		t.Errorf("Price info not recognized on %q", str)
	}
	t.Error(pi)
}
