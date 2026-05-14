package cantrace

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	MaxClassicDLC = 8
	MaxClassicID  = 0x7ff
	MaxExtendedID = 0x1fffffff
)

type Frame struct {
	ID        uint32
	Data      [MaxClassicDLC]byte
	DLC       uint8
	Flags     uint32
	Timestamp uint64
	BusID     string
}

type FlagDefinition struct {
	Mask uint32
	Name string
}

func NewFrame(id uint32, rawData []byte, explicitDLC *uint8, flags uint32) (Frame, error) {
	if id > MaxExtendedID {
		return Frame{}, fmt.Errorf("id 0x%X exceeds 29-bit CAN identifier range", id)
	}
	data, dlc, err := NormalizeData(rawData, explicitDLC)
	if err != nil {
		return Frame{}, err
	}
	return Frame{
		ID:    id,
		Data:  data,
		DLC:   dlc,
		Flags: flags,
	}, nil
}

func NormalizeData(rawData []byte, explicitDLC *uint8) ([MaxClassicDLC]byte, uint8, error) {
	if len(rawData) > MaxClassicDLC {
		return [MaxClassicDLC]byte{}, 0, fmt.Errorf("frame data exceeds %d bytes", MaxClassicDLC)
	}
	dlc := len(rawData)
	if explicitDLC != nil {
		if *explicitDLC > MaxClassicDLC {
			return [MaxClassicDLC]byte{}, 0, fmt.Errorf("dlc must be between 0 and %d", MaxClassicDLC)
		}
		if len(rawData) > int(*explicitDLC) {
			return [MaxClassicDLC]byte{}, 0, fmt.Errorf("data length %d exceeds dlc %d", len(rawData), *explicitDLC)
		}
		dlc = int(*explicitDLC)
	}
	var data [MaxClassicDLC]byte
	copy(data[:], rawData)
	return data, uint8(dlc), nil
}

func ParseDataBytes(raw string) ([]byte, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	fields := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ' ', '\t', ',', ':', ';', '-', '_':
			return true
		default:
			return false
		}
	})
	out := make([]byte, 0, len(fields))
	for _, field := range fields {
		token := strings.TrimSpace(field)
		token = strings.TrimPrefix(strings.ToLower(token), "0x")
		if token == "" {
			continue
		}
		n, err := strconv.ParseUint(token, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid data byte %q", field)
		}
		out = append(out, byte(n))
	}
	if len(out) > MaxClassicDLC {
		return nil, fmt.Errorf("frame data exceeds %d bytes", MaxClassicDLC)
	}
	return out, nil
}

func FrameData(frame Frame) []byte {
	dlc := frame.DLC
	if dlc > MaxClassicDLC {
		dlc = MaxClassicDLC
	}
	out := make([]byte, dlc)
	copy(out, frame.Data[:dlc])
	return out
}

func DataHex(data []byte) string {
	if len(data) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(data))
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, " ")
}

func FrameHex(frame Frame) string {
	return DataHex(FrameData(frame))
}

func InferFallbackDLC(data [MaxClassicDLC]byte) uint8 {
	n := MaxClassicDLC
	for n > 0 && data[n-1] == 0 {
		n--
	}
	return uint8(n)
}

func ResolveDLC(id uint32, data [MaxClassicDLC]byte, known map[uint32]uint8) uint8 {
	if dlc, ok := known[id]; ok && dlc <= MaxClassicDLC {
		return dlc
	}
	return InferFallbackDLC(data)
}

func ShouldSkipFlags(flags uint32, skipMask uint32) bool {
	return flags&skipMask != 0
}

func FlagNames(flags uint32, definitions []FlagDefinition) []string {
	out := make([]string, 0, len(definitions))
	for _, def := range definitions {
		if def.Mask != 0 && flags&def.Mask != 0 {
			out = append(out, def.Name)
		}
	}
	return out
}

func FormatFlagNames(flags uint32, definitions []FlagDefinition) string {
	return strings.Join(FlagNames(flags, definitions), ",")
}
