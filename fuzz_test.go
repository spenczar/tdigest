//go:build go1.18
// +build go1.18

package tdigest

import (
	"bytes"
	"testing"
)

// Past cases that revealed panics.
var fuzzFailures = [][]byte{
	[]byte{
		0x01, 0x00, 0x00, 0x00, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0xfc,
	},
	[]byte{
		0x01, 0x00, 0x00, 0x00, 0xdb, 0x46, 0x5f, 0xbd,
		0xdb, 0x46, 0x00, 0xbd, 0xe0, 0xdf, 0xca, 0xab,
		0x37, 0x31, 0x37, 0x32, 0x37, 0x33, 0x37, 0x34,
		0x37, 0x35, 0x37, 0x36, 0x37, 0x37, 0x37, 0x38,
		0x37, 0x39, 0x28,
	},
	[]byte{
		0x80, 0x0c, 0x01, 0x00, 0x00, 0x00, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x02, 0x00,
		0x00, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0xbf,
	},
	[]byte{
		0x80, 0x0c, 0x01, 0x00, 0x00, 0x00, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x02, 0x00,
		0x00, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x63, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x4e,
	},
	[]byte{
		0x80, 0x0c, 0x01, 0x00, 0x00, 0x00, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x02, 0x00,
		0x00, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x00, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x92, 0x00,
	},
}

func FuzzRoundTrip(f *testing.F) {

	for _, data := range fuzzFailures {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		v := new(TDigest)
		err := v.UnmarshalBinary(data)
		if err != nil {
			// Input is not valid; skip it.
			t.Skip()
		}

		t.Logf("input: %v", data)
		remarshaled, err := v.MarshalBinary()
		if err != nil {
			t.Fatalf("marshal error for valid data: %v", err)
		}

		if !bytes.HasPrefix(data, remarshaled) {
			t.Logf("tdigest: %s", v.debugStr())
			t.Fatal("remarshaling does not round-trip")
		}

		for q := float64(0.1); q <= 1.0; q += 0.05 {
			prev, this := v.Quantile(q-0.1), v.Quantile(q)
			if prev-this > 1e-100 { // Floating point math makes this slightly imprecise.
				t.Logf("tdigest: %s", v.debugStr())
				t.Logf("q: %v", q)
				t.Logf("prev: %v", prev)
				t.Logf("this: %v", this)
				t.Fatal("quantiles should only increase")
			}
		}
		v.Add(1, 1)

	})
}
