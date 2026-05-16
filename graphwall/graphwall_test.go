package graphwall

import (
	"testing"
	"time"
)

func TestCalculateViewport(t *testing.T) {
	// Base time: 2026-05-07 00:00:00 UTC
	base := time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC).UnixNano()

	tests := []struct {
		name          string
		req           TimeViewportRequest
		wantViewStart int64
		wantViewEnd   int64
		wantInterval  string
	}{
		{
			name: "Example 12:32 to 13:39",
			req: TimeViewportRequest{
				DataStartUnixNano:     base,
				DataEndUnixNano:       base + int64((13*3600+39*60)*time.Second),
				SelectedStartUnixNano: base + int64((12*3600+32*60)*time.Second),
			},
			wantViewStart: base + int64((12*3600+30*60)*time.Second),
			wantViewEnd:   base + int64((13*3600+45*60)*time.Second),
			wantInterval:  "15 min",
		},
		{
			name: "Very short range choose 1 minute",
			req: TimeViewportRequest{
				DataStartUnixNano:     base,
				DataEndUnixNano:       base + 10*int64(time.Second),
				SelectedStartUnixNano: base,
			},
			wantViewStart: base,
			wantViewEnd:   base + 60*int64(time.Second),
			wantInterval:  "1 min",
		},
		{
			name: "Exactly aligned range does not expand unnecessarily",
			req: TimeViewportRequest{
				DataStartUnixNano:     base,
				DataEndUnixNano:       base + 60*int64(time.Minute),
				SelectedStartUnixNano: base,
			},
			wantViewStart: base,
			wantViewEnd:   base + 60*int64(time.Minute),
			wantInterval:  "10 min", // 60 / 10 = 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateViewport(tt.req)
			if got.ViewStartUnixNano != tt.wantViewStart {
				t.Errorf("ViewStart = %v, want %v", time.Unix(0, got.ViewStartUnixNano), time.Unix(0, tt.wantViewStart))
			}
			if got.ViewEndUnixNano != tt.wantViewEnd {
				t.Errorf("ViewEnd = %v, want %v", time.Unix(0, got.ViewEndUnixNano), time.Unix(0, tt.wantViewEnd))
			}
			if got.Interval.Name != tt.wantInterval {
				t.Errorf("Interval = %s, want %s", got.Interval.Name, tt.wantInterval)
			}
		})
	}
}

func TestSemanticAggregateUsesCanonicalUnitAliases(t *testing.T) {
	tests := []struct {
		name string
		unit string
		want string
	}{
		{name: "watts", unit: "watts", want: AggregatePower},
		{name: "uppercase watt", unit: "W", want: AggregatePower},
		{name: "degrees celsius symbol", unit: "°C", want: AggregateTemperature},
		{name: "celsius word", unit: "celsius", want: AggregateTemperature},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SemanticAggregate(SemanticInput{Unit: tt.unit})
			if got != tt.want {
				t.Fatalf("SemanticAggregate(Unit: %q) = %q, want %q", tt.unit, got, tt.want)
			}
		})
	}
}
