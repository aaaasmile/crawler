package crawler

import "testing"

func TestParsePrice(t *testing.T) {
	//str := "IS.EO ST.SEL.DIV.30 U.ETF - Aktuell: 16,34 (15.01. / 17:36)"
	pricestr := "16,34"
	closed := "Schluss 15.01.23"
	pi, err := parseForPriceInfo(pricestr, closed)
	if err != nil {
		t.Errorf("Price info not recognized on %q with error:\n %v", pricestr, err)
		return
	}
	if pi.Price != 16.34 {
		t.Error("Expect 16.34 but recived ", pi.Price)
	}
	if pi.TimestampInt != 1610732160 {
		t.Error("Expect 1610732160 but recived ", pi.TimestampInt)
	}

}

func TestParsePrice2(t *testing.T) {
	str := "XTR.FTSE MIB 1D - Aktuell: 21,87 (22.01. / 17:36)"
	_, err := parseForPriceInfo(str)
	if err != nil {
		t.Errorf("Price info not recognized on %q with error:\n %v", str, err)
		return
	}
}

func TestParsePrice3(t *testing.T) {
	str := "ISHSVII-MSCI USA -SC DL AC - Aktuell: 374,10 (22.01. / 17:36)"
	pi, err := parseForPriceInfo(str)
	if err != nil {
		t.Errorf("Price info not recognized on %q with error:\n %v", str, err)
		return
	}

	if pi.Price != 374.10 {
		t.Error("Expect 374,10 but recived ", pi.Price)
		return
	}
}
