package dbcmeta

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ObservedMessage summarizes traffic observed for one CAN identifier.
type ObservedMessage struct {
	ID        uint32   `json:"id"`
	DLC       int      `json:"dlc,omitempty"`
	Count     int      `json:"count,omitempty"`
	SampleHex []string `json:"sample_hex,omitempty"`
}

// MatchResult ranks one catalogue candidate against observed CAN traffic.
type MatchResult struct {
	Representative     string   `json:"representative"`
	Aliases            []string `json:"aliases,omitempty"`
	FamilyKeys         []string `json:"family_keys,omitempty"`
	VersionKeys        []string `json:"version_keys,omitempty"`
	SemanticHash       string   `json:"semantic_hash"`
	Fingerprint        string   `json:"fingerprint"`
	MatchScore         float64  `json:"match_score"`
	ObservedCoverage   float64  `json:"observed_coverage"`
	DLCAgreement       float64  `json:"dlc_agreement"`
	IDPrecision        float64  `json:"id_precision"`
	SignalPlausibility float64  `json:"signal_plausibility"`
	MatchedObserved    int      `json:"matched_observed"`
	ObservedCount      int      `json:"observed_count"`
	CandidateMessages  int      `json:"candidate_messages"`
	CandidateSignals   int      `json:"candidate_signals"`
	Notes              []string `json:"notes,omitempty"`
}

// MessageFingerprint is a stable, inspectable summary of one DBC message.
type MessageFingerprint struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	DLC         int    `json:"dlc"`
	SignalCount int    `json:"signal_count"`
	Hash        string `json:"hash"`
	Signature   string `json:"signature"`
}

// CandidateMeta describes one canonical DBC candidate.
type CandidateMeta struct {
	Representative string   `json:"representative"`
	Path           string   `json:"path,omitempty"`
	Aliases        []string `json:"aliases,omitempty"`
	FamilyKeys     []string `json:"family_keys,omitempty"`
	VersionKeys    []string `json:"version_keys,omitempty"`
	SemanticHash   string   `json:"semantic_hash"`
	Fingerprint    string   `json:"fingerprint"`
	MessageCount   int      `json:"message_count"`
	SignalCount    int      `json:"signal_count"`
}

// CandidateDetail extends candidate metadata with per-message fingerprints.
type CandidateDetail struct {
	CandidateMeta
	MessageFingerprints []MessageFingerprint `json:"message_fingerprints,omitempty"`
}

// Catalog groups raw DBC files into canonical semantic candidates.
type Catalog struct {
	Kind                   string          `json:"kind"`
	Directory              string          `json:"directory,omitempty"`
	FileCount              int             `json:"file_count"`
	CanonicalCount         int             `json:"canonical_count"`
	RawDuplicateGroupCount int             `json:"raw_duplicate_group_count"`
	SemanticGroupCount     int             `json:"semantic_group_count"`
	Candidates             []CandidateMeta `json:"candidates,omitempty"`

	candidates []candidate
}

type candidate struct {
	CandidateMeta
	Messages            map[uint32]*Message
	MessageFingerprints []MessageFingerprint
}

type variantIdentity struct {
	NormalizedStem string
	FamilyKey      string
	VersionKey     string
	Timestamp      string
	CopyIndex      int
}

type loadedVariant struct {
	Name                string
	Path                string
	SHA256              string
	Identity            variantIdentity
	SemanticHash        string
	MessageCount        int
	SignalCount         int
	DB                  *File
	MessageFingerprints []MessageFingerprint
}

var (
	uploadPrefixRE = regexp.MustCompile(`^(\d{8}_\d{6})_(.+)$`)
	copySuffixRE   = regexp.MustCompile(`^(.*?)(?:\s*\((\d+)\)|_\((\d+)\))$`)
	versionRE      = regexp.MustCompile(`(?:^|_)(v?\d+(?:_\d+){1,3})(?:_(rc(?:_\d+|\d+)))?(?:$|_)`)
)

// LoadDir loads all .dbc files in a directory into a canonical catalogue.
func LoadDir(dir string) (*Catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".dbc" {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(paths)
	return LoadFiles(paths, dir)
}

// LoadFiles loads the given DBC files into a canonical catalogue.
func LoadFiles(paths []string, directory string) (*Catalog, error) {
	catalog := &Catalog{
		Kind:      "dbc_meta",
		Directory: directory,
		FileCount: len(paths),
	}
	if len(paths) == 0 {
		return catalog, nil
	}

	byRaw := make(map[string][]string)
	bySemantic := make(map[string][]loadedVariant)
	for _, path := range paths {
		variant, err := loadVariant(path)
		if err != nil {
			continue
		}
		byRaw[variant.SHA256] = append(byRaw[variant.SHA256], variant.Name)
		bySemantic[variant.SemanticHash] = append(bySemantic[variant.SemanticHash], variant)
	}

	catalog.RawDuplicateGroupCount = countStringGroups(byRaw)
	catalog.SemanticGroupCount = countVariantGroups(bySemantic)

	semanticKeys := make([]string, 0, len(bySemantic))
	for key := range bySemantic {
		semanticKeys = append(semanticKeys, key)
	}
	sort.Strings(semanticKeys)

	for _, key := range semanticKeys {
		variants := bySemantic[key]
		sort.Slice(variants, func(i, j int) bool { return variants[i].Name < variants[j].Name })

		aliasNames := make([]string, 0, len(variants))
		familyKeys := make(map[string]struct{})
		versionKeys := make(map[string]struct{})
		for _, variant := range variants {
			aliasNames = append(aliasNames, variant.Name)
			if variant.Identity.FamilyKey != "" {
				familyKeys[variant.Identity.FamilyKey] = struct{}{}
			}
			if variant.Identity.VersionKey != "" {
				versionKeys[variant.Identity.VersionKey] = struct{}{}
			}
		}

		rep := chooseRepresentativeVariant(variants)
		meta := CandidateMeta{
			Representative: rep.Name,
			Path:           rep.Path,
			Aliases:        aliasNames,
			FamilyKeys:     setToSortedList(familyKeys),
			VersionKeys:    sortVersions(setToSortedList(versionKeys)),
			SemanticHash:   key,
			Fingerprint:    shortHash(key),
			MessageCount:   variants[0].MessageCount,
			SignalCount:    variants[0].SignalCount,
		}
		catalog.Candidates = append(catalog.Candidates, meta)
		catalog.candidates = append(catalog.candidates, candidate{
			CandidateMeta:       meta,
			Messages:            variants[0].DB.Messages,
			MessageFingerprints: append([]MessageFingerprint(nil), variants[0].MessageFingerprints...),
		})
	}

	sort.Slice(catalog.Candidates, func(i, j int) bool {
		if catalog.Candidates[i].Representative == catalog.Candidates[j].Representative {
			return catalog.Candidates[i].SemanticHash < catalog.Candidates[j].SemanticHash
		}
		return catalog.Candidates[i].Representative < catalog.Candidates[j].Representative
	})
	sort.Slice(catalog.candidates, func(i, j int) bool {
		if catalog.candidates[i].Representative == catalog.candidates[j].Representative {
			return catalog.candidates[i].SemanticHash < catalog.candidates[j].SemanticHash
		}
		return catalog.candidates[i].Representative < catalog.candidates[j].Representative
	})
	catalog.CanonicalCount = len(catalog.Candidates)
	return catalog, nil
}

// DetailByKey looks up a candidate by semantic hash, fingerprint, representative, or alias.
func (c *Catalog) DetailByKey(key string) (CandidateDetail, bool) {
	if c == nil {
		return CandidateDetail{}, false
	}
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return CandidateDetail{}, false
	}
	for _, candidate := range c.candidates {
		if strings.ToLower(candidate.SemanticHash) == key ||
			strings.ToLower(candidate.Fingerprint) == key ||
			strings.ToLower(candidate.Representative) == key {
			return CandidateDetail{
				CandidateMeta:       candidate.CandidateMeta,
				MessageFingerprints: append([]MessageFingerprint(nil), candidate.MessageFingerprints...),
			}, true
		}
		for _, alias := range candidate.Aliases {
			if strings.ToLower(alias) == key {
				return CandidateDetail{
					CandidateMeta:       candidate.CandidateMeta,
					MessageFingerprints: append([]MessageFingerprint(nil), candidate.MessageFingerprints...),
				}, true
			}
		}
	}
	return CandidateDetail{}, false
}

// PathByKey resolves a candidate path by the same keys accepted by DetailByKey.
func (c *Catalog) PathByKey(key string) (string, bool) {
	detail, ok := c.DetailByKey(key)
	if !ok || strings.TrimSpace(detail.Path) == "" {
		return "", false
	}
	return detail.Path, true
}

// RankObserved ranks catalogue candidates against observed CAN traffic.
func (c *Catalog) RankObserved(observed []ObservedMessage, limit int) []MatchResult {
	if c == nil || len(c.candidates) == 0 || len(observed) == 0 {
		return nil
	}

	normalized := normalizeObserved(observed)
	totalWeight := 0
	for _, item := range normalized {
		totalWeight += observedWeight(item)
	}
	if totalWeight == 0 {
		return nil
	}

	results := make([]MatchResult, 0, len(c.candidates))
	for _, candidate := range c.candidates {
		matchedWeight := 0
		matchedCount := 0
		matchedDLCWeight := 0
		sampleWeight := 0
		sampleScoreWeight := 0.0

		for _, item := range normalized {
			msg, ok := candidate.Messages[item.ID]
			if !ok {
				continue
			}
			weight := observedWeight(item)
			matchedWeight += weight
			matchedCount++
			if item.DLC <= 0 || msg.DLC == item.DLC {
				matchedDLCWeight += weight
			}
			if score, ok := scoreMessagePlausibility(msg, item); ok {
				sampleWeight += weight
				sampleScoreWeight += score * float64(weight)
			}
		}

		if matchedCount == 0 {
			continue
		}

		coverage := float64(matchedWeight) / float64(totalWeight)
		dlcAgreement := float64(matchedDLCWeight) / float64(matchedWeight)
		idPrecision := float64(matchedCount) / float64(candidate.MessageCount)
		signalPlausibility := 0.5
		if sampleWeight > 0 {
			signalPlausibility = sampleScoreWeight / float64(sampleWeight)
		}
		score := (0.50 * coverage) + (0.15 * dlcAgreement) + (0.10 * idPrecision) + (0.25 * signalPlausibility)

		notes := make([]string, 0, 4)
		if coverage == 1 {
			notes = append(notes, "All observed frame IDs were found in this canonical DBC.")
		}
		if sampleWeight == 0 {
			notes = append(notes, "No payload samples were available, so specific implementation scoring is still mostly family-level.")
		} else {
			notes = append(notes, fmt.Sprintf("Signal plausibility used %d weighted payload sample group(s).", sampleWeight))
		}
		if dlcAgreement < 1 {
			notes = append(notes, "Some matched IDs disagree on DLC, so this may be a nearby branch rather than an exact fit.")
		}
		if len(candidate.VersionKeys) > 1 {
			notes = append(notes, "This semantic shape spans several version labels, so filename alone is not a reliable discriminator.")
		}

		results = append(results, MatchResult{
			Representative:     candidate.Representative,
			Aliases:            candidate.Aliases,
			FamilyKeys:         candidate.FamilyKeys,
			VersionKeys:        candidate.VersionKeys,
			SemanticHash:       candidate.SemanticHash,
			Fingerprint:        candidate.Fingerprint,
			MatchScore:         round3(score),
			ObservedCoverage:   round3(coverage),
			DLCAgreement:       round3(dlcAgreement),
			IDPrecision:        round3(idPrecision),
			SignalPlausibility: round3(signalPlausibility),
			MatchedObserved:    matchedCount,
			ObservedCount:      len(normalized),
			CandidateMessages:  candidate.MessageCount,
			CandidateSignals:   candidate.SignalCount,
			Notes:              notes,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].MatchScore == results[j].MatchScore {
			if results[i].SignalPlausibility == results[j].SignalPlausibility {
				return results[i].Representative < results[j].Representative
			}
			return results[i].SignalPlausibility > results[j].SignalPlausibility
		}
		return results[i].MatchScore > results[j].MatchScore
	})
	if limit > 0 && len(results) > limit {
		return results[:limit]
	}
	return results
}

func loadVariant(path string) (loadedVariant, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return loadedVariant{}, err
	}
	parsed, err := ParseFile(path)
	if err != nil {
		return loadedVariant{}, err
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	identity := parseIdentity(name)
	rawDigest := sha256.Sum256(data)
	rawHash := hex.EncodeToString(rawDigest[:])

	signalCount := 0
	for _, msg := range parsed.Messages {
		signalCount += len(msg.Signals)
	}

	return loadedVariant{
		Name:                name,
		Path:                path,
		SHA256:              rawHash,
		Identity:            identity,
		SemanticHash:        semanticSignature(parsed, data),
		MessageCount:        len(parsed.Messages),
		SignalCount:         signalCount,
		DB:                  parsed,
		MessageFingerprints: buildMessageFingerprints(parsed),
	}, nil
}

func buildMessageFingerprints(parsed *File) []MessageFingerprint {
	ids := sortedMessageIDs(parsed.Messages)
	out := make([]MessageFingerprint, 0, len(ids))
	for _, id := range ids {
		msg := parsed.Messages[id]
		signature := messageSignature(msg)
		hash := sha256.Sum256([]byte(signature))
		out = append(out, MessageFingerprint{
			ID:          msg.ID,
			Name:        msg.Name,
			DLC:         msg.DLC,
			SignalCount: len(msg.Signals),
			Hash:        shortHash(hex.EncodeToString(hash[:])),
			Signature:   signature,
		})
	}
	return out
}

func messageSignature(msg *Message) string {
	signals := append([]Signal(nil), msg.Signals...)
	sort.Slice(signals, func(i, j int) bool {
		if signals[i].StartBit == signals[j].StartBit {
			return strings.ToLower(signals[i].Name) < strings.ToLower(signals[j].Name)
		}
		return signals[i].StartBit < signals[j].StartBit
	})
	parts := make([]string, 0, len(signals))
	for _, sig := range signals {
		byteOrder := "le"
		if sig.BigEndian {
			byteOrder = "be"
		}
		signed := "u"
		if sig.Signed {
			signed = "s"
		}
		enumTag := ""
		if sig.IsBoolean {
			enumTag = " bool"
		} else if sig.IsEnum {
			enumTag = fmt.Sprintf(" enum=%d", len(sig.ValueTable))
		}
		parts = append(parts, fmt.Sprintf("%s@%d|%d %s%s f=%g o=%g [%g..%g]%s",
			sig.Name, sig.StartBit, sig.Length, byteOrder, signed, sig.Factor, sig.Offset, sig.Min, sig.Max, enumTag))
	}
	return fmt.Sprintf("0x%03X %s dlc=%d :: %s", msg.ID, msg.Name, msg.DLC, strings.Join(parts, "; "))
}

func semanticSignature(parsed *File, fallback []byte) string {
	var buf bytes.Buffer
	for _, id := range sortedMessageIDs(parsed.Messages) {
		msg := parsed.Messages[id]
		fmt.Fprintf(&buf, "BO_ %d %s %d %s\n", msg.ID, strings.ToLower(msg.Name), msg.DLC, strings.ToLower(msg.Sender))
		signals := append([]Signal(nil), msg.Signals...)
		sort.Slice(signals, func(i, j int) bool { return strings.ToLower(signals[i].Name) < strings.ToLower(signals[j].Name) })
		for _, sig := range signals {
			fmt.Fprintf(&buf, "SG_ %s %d %d %t %t %.12g %.12g %.12g %.12g %s bool=%t enum=%t val=%s\n",
				strings.ToLower(sig.Name),
				sig.StartBit,
				sig.Length,
				sig.BigEndian,
				sig.Signed,
				sig.Factor,
				sig.Offset,
				sig.Min,
				sig.Max,
				strings.ToLower(strings.TrimSpace(sig.Unit)),
				sig.IsBoolean,
				sig.IsEnum,
				valueTableSignature(sig.ValueTable),
			)
		}
	}
	sum := sha256.Sum256(buf.Bytes())
	if len(buf.Bytes()) == 0 {
		sum = sha256.Sum256(fallback)
	}
	return hex.EncodeToString(sum[:])
}

func valueTableSignature(table map[float64]string) string {
	if len(table) == 0 {
		return ""
	}
	keys := make([]float64, 0, len(table))
	for key := range table {
		keys = append(keys, key)
	}
	sort.Float64s(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%.12g=%s", key, strings.ToLower(strings.TrimSpace(table[key]))))
	}
	return strings.Join(parts, ",")
}

func sortedMessageIDs(messages map[uint32]*Message) []uint32 {
	ids := make([]uint32, 0, len(messages))
	for id := range messages {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func scoreMessagePlausibility(msg *Message, observed ObservedMessage) (float64, bool) {
	samples := normalizeSampleHex(observed.SampleHex, observed.DLC)
	if len(samples) == 0 {
		return 0, false
	}
	if len(msg.Signals) == 0 {
		return 0.5, true
	}

	total := 0.0
	for _, payload := range samples {
		total += samplePlausibility(msg, payload)
	}
	return total / float64(len(samples)), true
}

func samplePlausibility(msg *Message, payload []byte) float64 {
	if len(msg.Signals) == 0 {
		return 0.5
	}
	total := 0.0
	for i := range msg.Signals {
		total += signalPlausibility(&msg.Signals[i], DecodeSignal(payload, &msg.Signals[i]))
	}
	return total / float64(len(msg.Signals))
}

func signalPlausibility(sig *Signal, value float64) float64 {
	if sig == nil {
		return 0
	}
	if sig.IsBoolean {
		if near(value, 0, 1e-9) || near(value, 1, 1e-9) {
			return 1
		}
		return 0
	}
	if sig.IsEnum && len(sig.ValueTable) > 0 {
		rounded := float64(int64(value + 0.5))
		if _, ok := sig.ValueTable[rounded]; ok {
			return 1
		}
		return 0
	}
	if value >= sig.Min && value <= sig.Max {
		return 1
	}
	span := sig.Max - sig.Min
	if span <= 0 {
		if near(value, sig.Min, 1e-9) {
			return 1
		}
		return 0
	}
	if value < sig.Min {
		return maxFloat(0, 1-((sig.Min-value)/span))
	}
	return maxFloat(0, 1-((value-sig.Max)/span))
}

func normalizeObserved(observed []ObservedMessage) []ObservedMessage {
	byID := make(map[uint32]ObservedMessage)
	for _, item := range observed {
		if item.ID == 0 {
			continue
		}
		existing, ok := byID[item.ID]
		if !ok {
			byID[item.ID] = item
			continue
		}
		if existing.Count <= 0 && item.Count > 0 {
			existing.Count = item.Count
		} else if item.Count > 0 {
			existing.Count += item.Count
		}
		switch {
		case existing.DLC <= 0 && item.DLC > 0:
			existing.DLC = item.DLC
		case existing.DLC > 0 && item.DLC > 0 && existing.DLC != item.DLC:
			existing.DLC = 0
		}
		existing.SampleHex = mergeSamples(existing.SampleHex, item.SampleHex, 4)
		byID[item.ID] = existing
	}
	out := make([]ObservedMessage, 0, len(byID))
	for _, item := range byID {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func mergeSamples(left []string, right []string, limit int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(left)+len(right))
	for _, sample := range append(left, right...) {
		sample = normalizeHex(sample)
		if sample == "" {
			continue
		}
		if _, ok := seen[sample]; ok {
			continue
		}
		seen[sample] = struct{}{}
		out = append(out, sample)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func normalizeSampleHex(samples []string, dlc int) [][]byte {
	out := make([][]byte, 0, len(samples))
	for _, sample := range samples {
		sample = normalizeHex(sample)
		if sample == "" {
			continue
		}
		buf, err := hex.DecodeString(sample)
		if err != nil {
			continue
		}
		if dlc > 0 && len(buf) > dlc {
			buf = buf[:dlc]
		}
		out = append(out, buf)
	}
	return out
}

func normalizeHex(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.TrimPrefix(raw, "0x")
	raw = strings.ReplaceAll(raw, " ", "")
	return raw
}

func observedWeight(item ObservedMessage) int {
	if item.Count > 0 {
		return item.Count
	}
	return 1
}

func chooseRepresentativeVariant(variants []loadedVariant) loadedVariant {
	best := loadedVariant{}
	bestScore := int(^uint(0) >> 1)
	for _, variant := range variants {
		score := len(variant.Name)
		if variant.Identity.Timestamp != "" {
			score += 100
		}
		if variant.Identity.CopyIndex > 0 {
			score += 25
		}
		if score < bestScore || (score == bestScore && variant.Name < best.Name) {
			best = variant
			bestScore = score
		}
	}
	return best
}

func parseIdentity(name string) variantIdentity {
	identity := variantIdentity{}
	core := strings.TrimSpace(name)

	if match := uploadPrefixRE.FindStringSubmatch(core); len(match) == 3 {
		identity.Timestamp = match[1]
		core = match[2]
	}
	if match := copySuffixRE.FindStringSubmatch(core); len(match) == 4 {
		core = strings.TrimSpace(match[1])
		copyText := match[2]
		if copyText == "" {
			copyText = match[3]
		}
		if copyText != "" {
			if value, err := strconv.Atoi(copyText); err == nil {
				identity.CopyIndex = value
			}
		}
	}

	identity.NormalizedStem = normalizeStem(core)
	identity.FamilyKey, identity.VersionKey = splitFamilyAndVersion(identity.NormalizedStem)
	return identity
}

func normalizeStem(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer("(", " ", ")", " ", "_", " ", "-", " ", ".", " ")
	normalized = replacer.Replace(normalized)
	return strings.Join(strings.Fields(normalized), "_")
}

func splitFamilyAndVersion(normalized string) (string, string) {
	if normalized == "" {
		return "", ""
	}
	match := versionRE.FindStringSubmatch(normalized)
	if len(match) < 2 {
		return normalized, ""
	}
	pos := strings.Index(normalized, match[1])
	family := normalized
	if pos >= 0 {
		family = strings.TrimSuffix(strings.TrimSpace(normalized[:pos]), "_")
	}
	if family == "" {
		family = normalized
	}

	version := strings.TrimPrefix(match[1], "v")
	version = strings.ReplaceAll(version, "_", ".")
	if len(match) >= 3 && strings.TrimSpace(match[2]) != "" {
		stage := strings.TrimSpace(match[2])
		stage = strings.ReplaceAll(stage, "_", ".")
		version = version + "-" + stage
	}
	return family, version
}

func sortVersions(values []string) []string {
	out := append([]string(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		return compareVersion(parseVersion(out[i]), parseVersion(out[j])) < 0
	})
	return out
}

type version struct {
	Parts []int
	RC    *int
	OK    bool
	Raw   string
}

func parseVersion(raw string) version {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return version{}
	}
	v := version{Raw: raw}
	main := raw
	if idx := strings.Index(raw, "-"); idx >= 0 {
		main = raw[:idx]
		suffix := raw[idx+1:]
		if strings.HasPrefix(suffix, "rc.") {
			n, err := strconv.Atoi(strings.TrimPrefix(suffix, "rc."))
			if err == nil {
				v.RC = &n
			}
		}
	}
	for _, part := range strings.Split(main, ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return version{Raw: raw}
		}
		v.Parts = append(v.Parts, n)
	}
	v.OK = len(v.Parts) > 0
	return v
}

func compareVersion(left version, right version) int {
	if left.OK && !right.OK {
		return -1
	}
	if !left.OK && right.OK {
		return 1
	}
	if left.OK && right.OK {
		n := len(left.Parts)
		if len(right.Parts) > n {
			n = len(right.Parts)
		}
		for i := 0; i < n; i++ {
			lv := 0
			rv := 0
			if i < len(left.Parts) {
				lv = left.Parts[i]
			}
			if i < len(right.Parts) {
				rv = right.Parts[i]
			}
			if lv != rv {
				if lv < rv {
					return -1
				}
				return 1
			}
		}
		if left.RC == nil && right.RC != nil {
			return 1
		}
		if left.RC != nil && right.RC == nil {
			return -1
		}
		if left.RC != nil && right.RC != nil && *left.RC != *right.RC {
			if *left.RC < *right.RC {
				return -1
			}
			return 1
		}
	}
	return strings.Compare(left.Raw, right.Raw)
}

func setToSortedList(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func countStringGroups(values map[string][]string) int {
	count := 0
	for _, group := range values {
		if len(group) > 1 {
			count++
		}
	}
	return count
}

func countVariantGroups(values map[string][]loadedVariant) int {
	count := 0
	for _, group := range values {
		if len(group) > 1 {
			count++
		}
	}
	return count
}

func shortHash(value string) string {
	if len(value) > 12 {
		return value[:12]
	}
	return value
}

func near(left float64, right float64, epsilon float64) bool {
	if left > right {
		return left-right <= epsilon
	}
	return right-left <= epsilon
}

func round3(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}

func maxFloat(a float64, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
