package rasterize_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func expectNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func imageToPNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	expectNoErr(t, err)
	return buf.Bytes()
}
