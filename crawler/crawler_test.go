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
	if pi.Timestamp.Hour() != 17 {
		t.Error("Expect 17 but recived ", pi.Timestamp.Hour())
	}
	if pi.Timestamp.Hour() != 17 {
		t.Error("Expect 36 but recived ", pi.Timestamp.Minute())
	}
	if pi.TimestampInt != 1610732160 {
		t.Error("Expect 1610732160 but recived ", pi.TimestampInt)
	}

}
