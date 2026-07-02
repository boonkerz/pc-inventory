package policy

import "testing"

func TestLastNumber(t *testing.T) {
	cases := map[string]struct {
		val float64
		ok  bool
	}{
		"Dockercontainer Anzahl 7": {7, true},
		"42":                       {42, true},
		"frei 13.5%":               {13.5, true},
		"nur text":                 {0, false},
		"":                         {0, false},
		"a 3 b 9":                  {9, true},
	}
	for in, want := range cases {
		v, ok := lastNumber(in)
		if ok != want.ok || (ok && v != want.val) {
			t.Errorf("lastNumber(%q) = %v,%v erwartet %v,%v", in, v, ok, want.val, want.ok)
		}
	}
}

func TestCompare(t *testing.T) {
	if !compare(5, "<", 10) || compare(5, "<", 3) || !compare(5, ">=", 5) || !compare(5, "!=", 6) || compare(5, "==", 6) {
		t.Error("compare falsch")
	}
}
