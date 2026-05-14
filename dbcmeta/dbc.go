package dbcmeta

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Signal describes one DBC signal definition.
type Signal struct {
	Name       string
	StartBit   int
	Length     int
	BigEndian  bool
	Signed     bool
	Factor     float64
	Offset     float64
	Min        float64
	Max        float64
	Unit       string
	Receivers  []string
	ValueTable map[float64]string
	IsBoolean  bool
	IsEnum     bool
}

// Message describes one DBC CAN message definition.
type Message struct {
	ID      uint32
	Name    string
	DLC     int
	Sender  string
	Signals []Signal
}

// File is the parsed subset of a DBC file used by SignalForge metadata tools.
type File struct {
	Messages map[uint32]*Message
	ByName   map[string]*Message
}

var (
	reMsgLine = regexp.MustCompile(`^BO_\s+(\d+)\s+(\w+)\s*:\s*(\d+)\s+(\w+)`)
	reSigLine = regexp.MustCompile(`^\s*SG_\s+(\w+)\s*:\s*(\d+)\|(\d+)@([01])([+-])\s*` +
		`\(([^,]+),([^)]+)\)\s*\[([^|]+)\|([^\]]+)\]\s*"([^"]*)"\s*(.*)`)
	reValLine = regexp.MustCompile(`^VAL_\s+(\d+)\s+(\w+)\s+(.*);`)
	reValPair = regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s+"([^"]+)"`)
)

// ParseFile reads and parses a DBC file.
func ParseFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseBytes parses DBC content from memory.
func ParseBytes(data []byte) (*File, error) {
	return Parse(strings.NewReader(string(data)))
}

// Parse reads and parses DBC content. It intentionally supports the subset
// needed for metadata catalogues and CAN payload plausibility checks.
func Parse(r io.Reader) (*File, error) {
	db := &File{
		Messages: make(map[uint32]*Message),
		ByName:   make(map[string]*Message),
	}

	var curMsg *Message
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if m := reMsgLine.FindStringSubmatch(trimmed); m != nil {
			id, _ := strconv.ParseUint(m[1], 10, 32)
			dlc, _ := strconv.Atoi(m[3])
			msg := &Message{
				ID:     uint32(id),
				Name:   m[2],
				DLC:    dlc,
				Sender: m[4],
			}
			db.Messages[msg.ID] = msg
			db.ByName[msg.Name] = msg
			curMsg = msg
			continue
		}

		if curMsg != nil {
			if s := reSigLine.FindStringSubmatch(line); s != nil {
				sig, err := parseSignal(s)
				if err != nil {
					return nil, err
				}
				curMsg.Signals = append(curMsg.Signals, sig)
				continue
			}
		}

		if v := reValLine.FindStringSubmatch(trimmed); v != nil {
			id64, _ := strconv.ParseUint(v[1], 10, 32)
			applyValueTable(db, uint32(id64), v[2], v[3])
			continue
		}

		if trimmed == "" || (len(trimmed) > 0 && trimmed[0] != ' ' && trimmed[0] != '\t') {
			if !strings.HasPrefix(trimmed, "SG_") {
				curMsg = nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(db.Messages) == 0 {
		return nil, fmt.Errorf("no messages found in DBC content")
	}
	return db, nil
}

func parseSignal(parts []string) (Signal, error) {
	startBit, err := strconv.Atoi(parts[2])
	if err != nil {
		return Signal{}, fmt.Errorf("parse start bit for %s: %w", parts[1], err)
	}
	length, err := strconv.Atoi(parts[3])
	if err != nil {
		return Signal{}, fmt.Errorf("parse length for %s: %w", parts[1], err)
	}
	factor, err := strconv.ParseFloat(parts[6], 64)
	if err != nil {
		return Signal{}, fmt.Errorf("parse factor for %s: %w", parts[1], err)
	}
	offset, err := strconv.ParseFloat(parts[7], 64)
	if err != nil {
		return Signal{}, fmt.Errorf("parse offset for %s: %w", parts[1], err)
	}
	minValue, err := strconv.ParseFloat(parts[8], 64)
	if err != nil {
		return Signal{}, fmt.Errorf("parse minimum for %s: %w", parts[1], err)
	}
	maxValue, err := strconv.ParseFloat(parts[9], 64)
	if err != nil {
		return Signal{}, fmt.Errorf("parse maximum for %s: %w", parts[1], err)
	}

	var receivers []string
	for _, receiver := range strings.Split(parts[11], ",") {
		if receiver = strings.TrimSpace(receiver); receiver != "" {
			receivers = append(receivers, receiver)
		}
	}

	return Signal{
		Name:       parts[1],
		StartBit:   startBit,
		Length:     length,
		BigEndian:  parts[4] == "0",
		Signed:     parts[5] == "-",
		Factor:     factor,
		Offset:     offset,
		Min:        minValue,
		Max:        maxValue,
		Unit:       parts[10],
		Receivers:  receivers,
		IsBoolean:  length == 1,
		ValueTable: make(map[float64]string),
	}, nil
}

func applyValueTable(db *File, id uint32, sigName string, valPart string) {
	msg, ok := db.Messages[id]
	if !ok {
		return
	}
	for i := range msg.Signals {
		if msg.Signals[i].Name != sigName {
			continue
		}
		for _, pair := range reValPair.FindAllStringSubmatch(valPart, -1) {
			val, _ := strconv.ParseFloat(pair[1], 64)
			msg.Signals[i].ValueTable[val] = strings.TrimSpace(pair[2])
		}
		msg.Signals[i].IsEnum = len(msg.Signals[i].ValueTable) > 0
		if len(msg.Signals[i].ValueTable) == 2 {
			_, has0 := msg.Signals[i].ValueTable[0]
			_, has1 := msg.Signals[i].ValueTable[1]
			if has0 && has1 {
				msg.Signals[i].IsBoolean = true
				msg.Signals[i].IsEnum = false
			}
		}
		return
	}
}

// DecodeSignal returns the physical value of a signal from a CAN payload.
func DecodeSignal(payload []byte, sig *Signal) float64 {
	raw := extractRaw(payload, sig)

	var numeric float64
	if sig.Signed {
		msb := uint64(1) << uint(sig.Length-1)
		if raw&msb != 0 {
			raw |= ^uint64(0) << uint(sig.Length)
		}
		numeric = float64(int64(raw))
	} else {
		numeric = float64(raw)
	}

	return numeric*sig.Factor + sig.Offset
}

func extractRaw(payload []byte, sig *Signal) uint64 {
	var result uint64
	if sig == nil {
		return 0
	}
	if !sig.BigEndian {
		for i := 0; i < sig.Length; i++ {
			bitPos := sig.StartBit + i
			b := bitPos / 8
			bi := bitPos % 8
			if b < len(payload) && (payload[b]>>bi)&1 == 1 {
				result |= 1 << uint(i)
			}
		}
		return result
	}

	byteIdx := sig.StartBit / 8
	bitInByte := sig.StartBit % 8
	for i := 0; i < sig.Length; i++ {
		if byteIdx < len(payload) && (payload[byteIdx]>>bitInByte)&1 == 1 {
			result |= 1 << uint(sig.Length-1-i)
		}
		bitInByte--
		if bitInByte < 0 {
			bitInByte = 7
			byteIdx++
		}
	}
	return result
}
