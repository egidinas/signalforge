package pollqueue

import (
	"testing"
	"time"
)

type testItem struct {
	Key string
}

func testKey(item testItem) string {
	return item.Key
}

func TestQueueRotatesNormalItems(t *testing.T) {
	q := New[testItem, float64]([]testItem{{"a"}, {"b"}, {"c"}}, testKey)

	first := q.NextChunk(2)
	second := q.NextChunk(2)
	got := []string{first[0].Key, first[1].Key, second[0].Key, second[1].Key}
	want := []string{"a", "b", "c", "a"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("rotation = %#v, want %#v", got, want)
		}
	}
}

func TestQueueManualPollGoesToFrontOnce(t *testing.T) {
	normal := testItem{"normal"}
	manual := testItem{"manual"}
	q := New[testItem, float64]([]testItem{normal}, testKey)

	q.EnqueueFront(manual)
	q.EnqueueFront(manual)
	chunk := q.NextChunk(3)

	if len(chunk) != 2 {
		t.Fatalf("chunk length = %d, want 2", len(chunk))
	}
	if chunk[0].Key != manual.Key || chunk[1].Key != normal.Key {
		t.Fatalf("chunk = %#v", chunk)
	}
}

func TestQueueSustainedManualPollsDoNotStarveNormalRotation(t *testing.T) {
	normal := testItem{"normal"}
	q := New[testItem, float64]([]testItem{normal}, testKey)

	var got []string
	for i := 0; i < 6; i++ {
		q.EnqueueFront(testItem{Key: "manual-" + string(rune('a'+i))})
		chunk := q.NextChunk(1)
		if len(chunk) != 1 {
			t.Fatalf("chunk %d length = %d, want 1", i, len(chunk))
		}
		got = append(got, chunk[0].Key)
	}

	normalCount := 0
	for _, key := range got {
		if key == normal.Key {
			normalCount++
		}
	}
	if normalCount == 0 {
		t.Fatalf("normal item starved under sustained manual load: %#v", got)
	}
}

func TestQueueDoesNotEmitDuplicateKeyInOneChunk(t *testing.T) {
	item := testItem{"same"}
	q := New[testItem, float64]([]testItem{item}, testKey)
	q.EnqueueFront(item)

	chunk := q.NextChunk(2)
	if len(chunk) != 1 || chunk[0].Key != item.Key {
		t.Fatalf("chunk = %#v", chunk)
	}

	next := q.NextChunk(1)
	if len(next) != 1 || next[0].Key != item.Key {
		t.Fatalf("next = %#v", next)
	}
}

func TestQueueLatestIsSeededAndRecorded(t *testing.T) {
	item := testItem{"temperature"}
	q := New[testItem, float64]([]testItem{item}, testKey)

	seed, ok := q.Latest(item)
	if !ok {
		t.Fatal("missing seeded latest result")
	}
	if seed.Status != StatusNotSampled {
		t.Fatalf("seed status = %q", seed.Status)
	}

	observedAt := time.Unix(100, 0)
	q.Record(Result[testItem, float64]{Item: item, Value: 12.5, ObservedAt: observedAt})
	got, ok := q.Latest(item)
	if !ok {
		t.Fatal("missing latest result")
	}
	if got.Value != 12.5 || !got.ObservedAt.Equal(observedAt) || got.Status != "" {
		t.Fatalf("latest = %#v", got)
	}
}
