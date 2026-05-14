package cantrace

import "testing"

func TestParseDataBytesAcceptsCommonSeparators(t *testing.T) {
	got, err := ParseDataBytes("0x7a 12,00:7A-12_00;e2")
	if err != nil {
		t.Fatalf("ParseDataBytes: %v", err)
	}
	want := []byte{0x7a, 0x12, 0x00, 0x7a, 0x12, 0x00, 0xe2}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestNormalizeDataWithExplicitDLC(t *testing.T) {
	dlc := uint8(6)
	data, gotDLC, err := NormalizeData([]byte{1, 2, 3, 4}, &dlc)
	if err != nil {
		t.Fatalf("NormalizeData: %v", err)
	}
	if gotDLC != 6 {
		t.Fatalf("dlc = %d, want 6", gotDLC)
	}
	if data[0] != 1 || data[3] != 4 || data[4] != 0 {
		t.Fatalf("data not copied/padded as expected: %v", data)
	}
}

func TestNormalizeDataRejectsDataLongerThanDLC(t *testing.T) {
	dlc := uint8(2)
	if _, _, err := NormalizeData([]byte{1, 2, 3}, &dlc); err == nil {
		t.Fatal("expected error for data longer than explicit dlc")
	}
}

func TestInferAndResolveDLC(t *testing.T) {
	var data [MaxClassicDLC]byte
	copy(data[:], []byte{0x7d, 0x00, 0x7d, 0x00, 0x08, 0x00, 0x00, 0x00})
	if got := InferFallbackDLC(data); got != 5 {
		t.Fatalf("fallback dlc = %d, want 5", got)
	}
	known := map[uint32]uint8{9: 4}
	if got := ResolveDLC(9, data, known); got != 4 {
		t.Fatalf("resolved dlc = %d, want 4", got)
	}
}

func TestNewFrameAndFrameHex(t *testing.T) {
	frame, err := NewFrame(0x123, []byte{0xAA, 0x00, 0x01}, nil, 0)
	if err != nil {
		t.Fatalf("NewFrame: %v", err)
	}
	if frame.DLC != 3 {
		t.Fatalf("dlc = %d, want 3", frame.DLC)
	}
	if got := FrameHex(frame); got != "AA 00 01" {
		t.Fatalf("FrameHex = %q", got)
	}
}

func TestFlags(t *testing.T) {
	const (
		ext   = 1 << 0
		txack = 1 << 1
		errf  = 1 << 2
	)
	defs := []FlagDefinition{
		{Mask: ext, Name: "EXT"},
		{Mask: txack, Name: "TXACK"},
		{Mask: errf, Name: "ERROR"},
	}
	if !ShouldSkipFlags(txack, txack|errf) {
		t.Fatal("expected txack to be skipped")
	}
	if got := FormatFlagNames(ext|errf, defs); got != "EXT,ERROR" {
		t.Fatalf("FormatFlagNames = %q", got)
	}
}
