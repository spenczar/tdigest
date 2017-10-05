package tdigest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const encodingVersion = int32(1)

func marshalBinary(d *TDigest) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := &binaryBufferWriter{buf: buf}
	w.writeValue(encodingVersion)

	var min, max float64
	if len(d.centroids) > 0 {
		min = d.centroids[0].mean
		max = d.centroids[len(d.centroids)-1].mean
	}

	w.writeValue(min)
	w.writeValue(max)
	w.writeValue(d.compression)
	w.writeValue(int32(len(d.centroids)))
	for _, c := range d.centroids {
		w.writeValue(c.count)
		w.writeValue(c.mean)
	}

	if w.err != nil {
		return nil, w.err
	}
	return buf.Bytes(), nil
}

func unmarshalBinary(d *TDigest, p []byte) error {
	var (
		ev       int32
		min, max float64
		n        int32
	)
	r := &binaryReader{r: bytes.NewReader(p)}
	r.readValue(&ev)
	if ev != encodingVersion {
		return fmt.Errorf("invalid encoding version: %d", ev)
	}
	r.readValue(&min)
	r.readValue(&max)
	r.readValue(&d.compression)
	r.readValue(&n)
	if r.err != nil {
		return r.err
	}
	d.centroids = make([]*centroid, int(n))
	for i := 0; i < int(n); i++ {
		c := new(centroid)
		r.readValue(&c.count)
		r.readValue(&c.mean)
		if r.err != nil {
			return r.err
		}
		d.centroids[i] = c
		d.countTotal += c.count
	}

	return r.err
}

type binaryBufferWriter struct {
	buf *bytes.Buffer
	err error
}

func (w *binaryBufferWriter) writeValue(v interface{}) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.buf, binary.LittleEndian, v)
}

type binaryReader struct {
	r   io.Reader
	err error
}

func (r *binaryReader) readValue(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, binary.LittleEndian, v)
}
