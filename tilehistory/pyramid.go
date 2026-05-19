package tilehistory

import (
	"time"
)

// Level represents a data pyramid tier with a specific cadence and retention.
type Level struct {
	Interval  time.Duration
	Retention time.Duration
	History   *History[float64]
}

// Pyramid manages multiple tiers of bucketed history (Google Maps like LOD).
type Pyramid struct {
	Levels []Level
}

func NewPyramid(specs ...LevelSpec) *Pyramid {
	p := &Pyramid{
		Levels: make([]Level, 0, len(specs)),
	}
	for _, spec := range specs {
		p.Levels = append(p.Levels, Level{
			Interval:  spec.Interval,
			Retention: spec.Retention,
			History:   NewWithInterval[float64](spec.Interval, spec.Retention),
		})
	}
	return p
}

type LevelSpec struct {
	Interval  time.Duration
	Retention time.Duration
}

func (p *Pyramid) Add(t time.Time, v float64) {
	sample := Sample[float64]{Timestamp: t, Value: v}
	for _, l := range p.Levels {
		l.History.Add(sample)
	}
}

func (p *Pyramid) Snapshot(interval time.Duration) Snapshot {
	// Pick the best level for the requested interval
	var best *Level
	for i := range p.Levels {
		if p.Levels[i].Interval <= interval {
			if best == nil || p.Levels[i].Interval > best.Interval {
				best = &p.Levels[i]
			}
		}
	}
	if best == nil && len(p.Levels) > 0 {
		best = &p.Levels[0]
	}
	if best == nil {
		return Snapshot{}
	}
	return best.History.Snapshot()
}

func (p *Pyramid) Earliest() time.Time {
	var earliest time.Time
	for _, l := range p.Levels {
		snap := l.History.Snapshot()
		if !snap.Earliest.IsZero() && (earliest.IsZero() || snap.Earliest.Before(earliest)) {
			earliest = snap.Earliest
		}
	}
	return earliest
}

func (p *Pyramid) Latest() time.Time {
	var latest time.Time
	for _, l := range p.Levels {
		snap := l.History.Snapshot()
		if !snap.Latest.IsZero() && (latest.IsZero() || snap.Latest.After(latest)) {
			latest = snap.Latest
		}
	}
	return latest
}
