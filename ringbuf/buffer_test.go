package ringbuf

import (
	"reflect"
	"testing"
)

type record struct {
	Name string
}

func recordSize(r record) int {
	return 10 + len(r.Name)
}

func TestBufferEvictsOldestFirst(t *testing.T) {
	buf := New[record](recordSize(record{"a"})*2, recordSize)

	first := record{"a"}
	second := record{"b"}
	third := record{"c"}

	buf.Push(first)
	buf.Push(second)
	buf.Push(third)

	got := buf.Drain()
	want := []record{second, third}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Drain = %#v, want %#v", got, want)
	}
}

func TestBufferPreservesInsertionOrder(t *testing.T) {
	buf := New[record](100, recordSize)
	rows := []record{{"a"}, {"b"}, {"c"}}

	for _, row := range rows {
		if !buf.Push(row) {
			t.Fatalf("Push rejected %#v", row)
		}
	}

	got := buf.Drain()
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("Drain = %#v, want %#v", got, rows)
	}
}

func TestBufferDrainResetsBuffer(t *testing.T) {
	buf := New[record](100, recordSize)
	first := record{"a"}
	second := record{"b"}

	buf.Push(first)
	if got := buf.Len(); got != 1 {
		t.Fatalf("Len before drain = %d, want 1", got)
	}
	if got := buf.Drain(); !reflect.DeepEqual(got, []record{first}) {
		t.Fatalf("first Drain = %#v", got)
	}
	if got := buf.Len(); got != 0 {
		t.Fatalf("Len after drain = %d, want 0", got)
	}
	if got := buf.Bytes(); got != 0 {
		t.Fatalf("Bytes after drain = %d, want 0", got)
	}

	buf.Push(second)
	if got := buf.Drain(); !reflect.DeepEqual(got, []record{second}) {
		t.Fatalf("second Drain = %#v", got)
	}
}

func TestBufferByteAccountingAndOversizeDrop(t *testing.T) {
	first := record{"alpha"}
	second := record{"beta"}
	buf := New[record](recordSize(first)+recordSize(second), recordSize)

	if !buf.Push(first) || !buf.Push(second) {
		t.Fatal("expected both records to fit")
	}
	if got := buf.Bytes(); got != recordSize(first)+recordSize(second) {
		t.Fatalf("Bytes = %d, want %d", got, recordSize(first)+recordSize(second))
	}
	if buf.Push(record{"this-record-is-too-large"}) {
		t.Fatal("oversize record was accepted")
	}
	if got := buf.Snapshot(); !reflect.DeepEqual(got, []record{first, second}) {
		t.Fatalf("Snapshot after oversize = %#v", got)
	}
}

func TestBufferGrowthAfterWraparound(t *testing.T) {
	buf := New[record](recordSize(record{"x"})*3+recordSize(record{"larger"}), recordSize)

	buf.Push(record{"a"})
	buf.Push(record{"b"})
	buf.Push(record{"c"})
	buf.Push(record{"d"})
	large := record{"larger"}
	buf.Push(large)

	got := buf.Snapshot()
	if len(got) == 0 {
		t.Fatal("Snapshot returned no records")
	}
	if got[len(got)-1] != large {
		t.Fatalf("last record = %#v, want %#v", got[len(got)-1], large)
	}
}
