package graphsem

import (
	"fmt"
	"sort"
	"strings"
)

const CurrentSourceProjectionSchemaVersion = 1

type ProjectionMappingKind string

const (
	ProjectionPrimary   ProjectionMappingKind = "primary"
	ProjectionSecondary ProjectionMappingKind = "secondary"
	ProjectionHidden    ProjectionMappingKind = "hidden"
	ProjectionPreview   ProjectionMappingKind = "preview"
)

type SignalProjectionBundle struct {
	SchemaVersion   int                       `json:"schema_version"`
	Namespace       string                    `json:"namespace,omitempty"`
	Source          string                    `json:"source,omitempty"`
	DefaultGrouping string                    `json:"default_grouping,omitempty"`
	Mappings        []SignalProjectionMapping `json:"mappings"`
	Metadata        map[string]string         `json:"metadata,omitempty"`
}

type SignalProjectionMapping struct {
	ID               string                        `json:"id,omitempty"`
	Kind             ProjectionMappingKind         `json:"kind"`
	Path             []SignalProjectionPathSegment `json:"path"`
	SignalRefs       []SignalProjectionRef         `json:"signal_refs"`
	GroupKey         string                        `json:"group_key,omitempty"`
	DeviceGrouping   string                        `json:"device_grouping,omitempty"`
	SortKey          string                        `json:"sort_key,omitempty"`
	DefaultVisible   bool                          `json:"default_visible,omitempty"`
	DefaultCollapsed bool                          `json:"default_collapsed,omitempty"`
	Reason           string                        `json:"reason,omitempty"`
	SourceID         string                        `json:"source_id,omitempty"`
	SourceFamily     SourceFamily                  `json:"source_family,omitempty"`
	Title            string                        `json:"title,omitempty"`
	Description      string                        `json:"description,omitempty"`
	Metadata         map[string]string             `json:"metadata,omitempty"`
}

type SignalProjectionPathSegment struct {
	ID               string            `json:"id"`
	Label            string            `json:"label"`
	Kind             string            `json:"kind,omitempty"`
	Order            int               `json:"order,omitempty"`
	DefaultCollapsed bool              `json:"default_collapsed,omitempty"`
	Description      string            `json:"description,omitempty"`
	SourceFamily     SourceFamily      `json:"source_family,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

type SignalProjectionRef struct {
	SignalID     SignalID          `json:"signal_id,omitempty"`
	SID          string            `json:"sid,omitempty"`
	TraceID      string            `json:"trace_id,omitempty"`
	SourceID     string            `json:"source_id,omitempty"`
	SourceFamily SourceFamily      `json:"source_family,omitempty"`
	TargetID     string            `json:"target_id,omitempty"`
	Role         SignalRole        `json:"role,omitempty"`
	Unit         string            `json:"unit,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type SignalProjectionValidationOptions struct {
	Catalogues               []SourceCatalogue
	AvailableSignalIDs       []SignalID
	RequiredPrimarySignalIDs []SignalID
	AllowPrimaryDuplicates   bool
	AllowUnknownSourceRefs   bool
	SkipSecondaryReasonCheck bool
}

type SignalProjectionTreeNode struct {
	ID               string                        `json:"id"`
	Label            string                        `json:"label"`
	Kind             string                        `json:"kind,omitempty"`
	Path             []SignalProjectionPathSegment `json:"path"`
	Order            int                           `json:"order,omitempty"`
	DefaultCollapsed bool                          `json:"default_collapsed,omitempty"`
	Mappings         []SignalProjectionMapping     `json:"mappings,omitempty"`
	Children         []SignalProjectionTreeNode    `json:"children,omitempty"`
}

func SignalProjectionRefKey(ref SignalProjectionRef) string {
	if ref.SignalID != "" {
		return "id:" + string(ref.SignalID)
	}
	if strings.TrimSpace(ref.SID) != "" {
		return "sid:" + strings.TrimSpace(ref.SID)
	}
	if strings.TrimSpace(ref.TraceID) != "" {
		return "trace:" + string(ref.SourceFamily) + ":" + strings.TrimSpace(ref.SourceID) + ":" + strings.TrimSpace(ref.TraceID)
	}
	if strings.TrimSpace(ref.TargetID) != "" {
		return "target:" + strings.TrimSpace(ref.TargetID)
	}
	return ""
}

func SignalProjectionPathKey(path []SignalProjectionPathSegment) string {
	ids := make([]string, 0, len(path))
	for _, segment := range path {
		ids = append(ids, segment.ID)
	}
	return strings.Join(ids, "/")
}

func ValidateSignalProjectionBundle(bundle SignalProjectionBundle, options SignalProjectionValidationOptions) error {
	if bundle.SchemaVersion != CurrentSourceProjectionSchemaVersion {
		return fmt.Errorf("signal projection schema_version must be %d, got %d", CurrentSourceProjectionSchemaVersion, bundle.SchemaVersion)
	}
	availableSignalIDs := map[string]struct{}{}
	for _, id := range options.AvailableSignalIDs {
		availableSignalIDs["id:"+string(id)] = struct{}{}
	}
	availableTraceRefs := map[string]struct{}{}
	for _, catalogue := range options.Catalogues {
		if err := catalogue.Validate(); err != nil {
			return fmt.Errorf("source catalogue %q invalid: %w", catalogue.SourceID, err)
		}
		for _, entry := range catalogue.Entries {
			addProjectionTraceKeys(availableTraceRefs, catalogue.SourceFamily, catalogue.SourceID, entry.TraceID)
		}
	}

	requiredPrimary := map[string]struct{}{}
	for _, id := range options.RequiredPrimarySignalIDs {
		requiredPrimary["id:"+string(id)] = struct{}{}
	}
	primarySeen := map[string]string{}
	for i, mapping := range bundle.Mappings {
		mappingID := mapping.ID
		if mappingID == "" {
			mappingID = fmt.Sprintf("mapping[%d]", i)
		}
		if mapping.Kind == "" {
			return fmt.Errorf("%s kind is required", mappingID)
		}
		if len(mapping.Path) == 0 {
			return fmt.Errorf("%s path is required", mappingID)
		}
		for j, segment := range mapping.Path {
			if strings.TrimSpace(segment.ID) == "" {
				return fmt.Errorf("%s path[%d] id is required", mappingID, j)
			}
			if strings.TrimSpace(segment.Label) == "" {
				return fmt.Errorf("%s path[%d] label is required", mappingID, j)
			}
		}
		if len(mapping.SignalRefs) == 0 {
			return fmt.Errorf("%s signal_refs is required", mappingID)
		}
		if mapping.Kind == ProjectionSecondary && !options.SkipSecondaryReasonCheck && len(strings.TrimSpace(mapping.Reason)) < 12 {
			return fmt.Errorf("%s secondary projection requires a review reason", mappingID)
		}
		for _, ref := range mapping.SignalRefs {
			key := SignalProjectionRefKey(ref)
			if key == "" {
				return fmt.Errorf("%s contains an empty signal reference", mappingID)
			}
			if strings.HasPrefix(key, "id:") && len(availableSignalIDs) > 0 {
				if _, ok := availableSignalIDs[key]; !ok {
					return fmt.Errorf("%s maps unknown signal %s", mappingID, key)
				}
			}
			if strings.HasPrefix(key, "trace:") && len(availableTraceRefs) > 0 && !options.AllowUnknownSourceRefs {
				if _, ok := availableTraceRefs[key]; !ok {
					return fmt.Errorf("%s maps unknown source trace %s", mappingID, key)
				}
			}
			if mapping.Kind == ProjectionPrimary {
				delete(requiredPrimary, key)
				if prior, ok := primarySeen[key]; ok && !options.AllowPrimaryDuplicates {
					return fmt.Errorf("%s duplicates primary projection for %s already mapped by %s", mappingID, key, prior)
				}
				primarySeen[key] = mappingID
			}
		}
	}
	if len(requiredPrimary) > 0 {
		missing := make([]string, 0, len(requiredPrimary))
		for key := range requiredPrimary {
			missing = append(missing, key)
		}
		sort.Strings(missing)
		return fmt.Errorf("missing primary projections: %s", strings.Join(missing, ", "))
	}
	return nil
}

func BuildSignalProjectionTree(bundle SignalProjectionBundle) []SignalProjectionTreeNode {
	roots := map[string]*SignalProjectionTreeNode{}
	for _, mapping := range bundle.Mappings {
		var node *SignalProjectionTreeNode
		path := []SignalProjectionPathSegment{}
		for depth, segment := range mapping.Path {
			path = append(path, segment)
			if depth == 0 {
				node = roots[segment.ID]
				if node == nil {
					node = projectionNode(segment, path)
					roots[segment.ID] = node
				}
				continue
			}
			node = ensureProjectionChild(node, segment, path)
		}
		if node != nil {
			node.Mappings = append(node.Mappings, mapping)
		}
	}
	return sortedProjectionNodes(roots)
}

func addProjectionTraceKeys(out map[string]struct{}, family SourceFamily, sourceID string, traceID string) {
	traceID = strings.TrimSpace(traceID)
	sourceID = strings.TrimSpace(sourceID)
	if traceID == "" {
		return
	}
	keys := []string{
		"trace:" + string(family) + ":" + sourceID + ":" + traceID,
		"trace::" + sourceID + ":" + traceID,
		"trace:" + string(family) + "::" + traceID,
		"trace:::" + traceID,
	}
	for _, key := range keys {
		out[key] = struct{}{}
	}
}

func projectionNode(segment SignalProjectionPathSegment, path []SignalProjectionPathSegment) *SignalProjectionTreeNode {
	return &SignalProjectionTreeNode{
		ID:               segment.ID,
		Label:            segment.Label,
		Kind:             segment.Kind,
		Path:             append([]SignalProjectionPathSegment(nil), path...),
		Order:            segment.Order,
		DefaultCollapsed: segment.DefaultCollapsed,
	}
}

func ensureProjectionChild(parent *SignalProjectionTreeNode, segment SignalProjectionPathSegment, path []SignalProjectionPathSegment) *SignalProjectionTreeNode {
	if parent == nil {
		return nil
	}
	for i := range parent.Children {
		if parent.Children[i].ID == segment.ID {
			return &parent.Children[i]
		}
	}
	child := projectionNode(segment, path)
	parent.Children = append(parent.Children, *child)
	return &parent.Children[len(parent.Children)-1]
}

func sortedProjectionNodes(lookup map[string]*SignalProjectionTreeNode) []SignalProjectionTreeNode {
	nodes := make([]SignalProjectionTreeNode, 0, len(lookup))
	for _, node := range lookup {
		copyNode := *node
		childLookup := map[string]*SignalProjectionTreeNode{}
		for i := range copyNode.Children {
			child := &copyNode.Children[i]
			childLookup[child.ID] = child
		}
		copyNode.Children = sortedProjectionNodes(childLookup)
		nodes = append(nodes, copyNode)
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Order != nodes[j].Order {
			if nodes[i].Order == 0 {
				return false
			}
			if nodes[j].Order == 0 {
				return true
			}
			return nodes[i].Order < nodes[j].Order
		}
		if nodes[i].Label != nodes[j].Label {
			return nodes[i].Label < nodes[j].Label
		}
		return nodes[i].ID < nodes[j].ID
	})
	return nodes
}
