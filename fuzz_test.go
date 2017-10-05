package tdigest

import "testing"

func TestFuzzPanicRegressions(t *testing.T) {
	// This test contains a list of byte sequences discovered by
	// github.com/dvyukov/go-fuzz which, at one time, caused tdigest to panic. The
	// test just makes sure that they no longer cause a panic.
	testcase := func(crasher []byte) func(*testing.T) {
		return func(*testing.T) {
			v := new(TDigest)
			err := v.UnmarshalBinary(crasher)
			if err != nil {
				return
			}
			_, err = v.MarshalBinary()
			if err != nil {
				panic(err)
			}
		}
	}
	t.Run("fuzz1", testcase([]byte(
		"\x01\x00\x00\x000000000000000000"+
			"00000000000\xfc")))
	t.Run("fuzz2", testcase([]byte(
		"\x01\x00\x00\x00\xdbF_\xbd\xdbF\x00\xbd\xe0\xdfʫ7172"+
			"73747576777879(")))
}
