package tdigest

import (
	"reflect"
	"testing"
)

func TestMarshalRoundTrip(t *testing.T) {
	testcase := func(in *TDigest) func(*testing.T) {
		return func(t *testing.T) {
			b, err := in.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary err: %v", err)
			}
			out := new(TDigest)
			err = out.UnmarshalBinary(b)
			if err != nil {
				t.Fatalf("UnmarshalBinary err: %v", err)
			}
			if !reflect.DeepEqual(in, out) {
				t.Errorf("marshaling round trip resulted in changes")
				t.Logf("in: %+v", in)
				t.Logf("out: %+v", out)
			}
		}
	}
	t.Run("empty", testcase(New()))
	t.Run("1 value", testcase(simpleTDigest(1)))
	t.Run("1000 values", testcase(simpleTDigest(1000)))
}
