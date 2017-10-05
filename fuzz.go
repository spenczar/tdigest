// +build gofuzz

package tdigest

func Fuzz(data []byte) int {
	v := new(TDigest)
	err := v.UnmarshalBinary(data)
	if err != nil {
		return 0
	}
	_, err = v.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return 1

}
