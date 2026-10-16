package linear

import (
	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/internal/common/names"
	label_repo "github.com/unbindapp/zigzag-api/internal/repositories/label"
)

const defaultLabelColor = "#95a2b3"

type labelTarget struct {
	id           uuid.UUID
	name         string
	teamID       *uuid.UUID
	parentID     *uuid.UUID
	isGroup      bool
	singleSelect bool
}

type labelScope struct {
	teamID   uuid.UUID
	parentID uuid.UUID
}

func scopeOf(teamID, parentID *uuid.UUID) labelScope {
	scope := labelScope{}
	if teamID != nil {
		scope.teamID = *teamID
	}
	if parentID != nil {
		scope.parentID = *parentID
	}
	return scope
}

func (self *planner) planLabels() {
	siblings := map[labelScope]map[string]*labelTarget{}
	register := func(target *labelTarget) {
		scope := scopeOf(target.teamID, target.parentID)
		if siblings[scope] == nil {
			siblings[scope] = map[string]*labelTarget{}
		}
		siblings[scope][lower(target.name)] = target
		self.labelsByName[lower(target.name)] = append(self.labelsByName[lower(target.name)], target)
		self.labelsByID[target.id] = target
	}
	for _, l := range self.snapshot.Labels {
		register(&labelTarget{id: l.ID, name: l.Name, teamID: l.TeamID, parentID: l.ParentID, isGroup: l.IsGroup, singleSelect: l.GroupSingleSelect})
	}

	parents := map[string]bool{}
	for _, l := range self.export.Labels {
		if id := text(l.ParentID); id != "" {
			parents[id] = true
		}
	}
	ordered := make([]LinearLabel, 0, len(self.export.Labels))
	for _, l := range self.export.Labels {
		if text(l.ParentID) == "" {
			ordered = append(ordered, l)
		}
	}
	for _, l := range self.export.Labels {
		if text(l.ParentID) != "" {
			ordered = append(ordered, l)
		}
	}

	for _, l := range ordered {
		var teamID *uuid.UUID
		if linearTeam := text(l.TeamID); linearTeam != "" {
			team, ok := self.teams[linearTeam]
			if !ok {
				self.warn(EntityLabels, l.ID, "label %q belongs to a team that is not imported", l.Name)
				self.count(EntityLabels, skipped)
				continue
			}
			teamID = &team.id
		}
		var parentID *uuid.UUID
		singleSelect := false
		if linearParent := text(l.ParentID); linearParent != "" {
			parent, ok := self.labels[linearParent]
			if !ok {
				self.warn(EntityLabels, l.ID, "label %q belongs to a group that is not imported", l.Name)
				self.count(EntityLabels, skipped)
				continue
			}
			parentID = &parent.id
		}
		if l.IsGroup {
			singleSelect = text(l.GroupType) == "singleSelect"
		}

		scope := scopeOf(teamID, parentID)
		if existing, ok := siblings[scope][lower(l.Name)]; ok {
			self.labels[l.ID] = existing
			self.count(EntityLabels, skipped)
			continue
		}
		name, err := names.Clean(l.Name)
		if err != nil {
			self.warn(EntityLabels, l.ID, "label has no name and was skipped")
			self.count(EntityLabels, skipped)
			continue
		}
		name = clip(name, maxNameLength)
		target := &labelTarget{id: uuid.New(), name: name, teamID: teamID, parentID: parentID, isGroup: l.IsGroup || parents[l.ID], singleSelect: singleSelect}
		register(target)
		self.labels[l.ID] = target
		self.plan.Labels = append(self.plan.Labels, label_repo.ImportedLabel{
			ID:                target.id,
			TeamID:            teamID,
			ParentID:          parentID,
			Name:              name,
			Color:             validColor(text(l.Color), defaultLabelColor),
			Description:       text(l.Description),
			IsGroup:           target.isGroup,
			GroupSingleSelect: singleSelect,
			SortOrder:         float64(len(siblings[scope]) - 1),
			ArchivedAt:        utc(l.ArchivedAt),
		})
		self.count(EntityLabels, created)
	}
}

func (self *planner) issueLabels(issue LinearIssue, teamID uuid.UUID) []uuid.UUID {
	var resolved []*labelTarget
	for _, id := range issue.LabelIDs {
		if target, ok := self.labels[id]; ok {
			resolved = append(resolved, target)
		}
	}
	if len(resolved) < len(issue.LabelIDs) {
		for _, name := range issue.LabelNames {
			if target := self.labelNamed(name, teamID); target != nil {
				resolved = append(resolved, target)
			}
		}
	}

	ids := []uuid.UUID{}
	groups := map[uuid.UUID]bool{}
	seen := map[uuid.UUID]bool{}
	for _, target := range resolved {
		if seen[target.id] {
			continue
		}
		seen[target.id] = true
		if target.isGroup {
			self.warn(EntityIssues, issue.Identifier, "label group %q cannot be applied to an issue and was left out", target.name)
			continue
		}
		if target.teamID != nil && *target.teamID != teamID {
			self.warn(EntityIssues, issue.Identifier, "label %q belongs to another team and was left out", target.name)
			continue
		}
		if target.parentID != nil && self.labelsByID[*target.parentID] != nil && self.labelsByID[*target.parentID].singleSelect {
			if groups[*target.parentID] {
				self.warn(EntityIssues, issue.Identifier, "label %q is a second choice in a single-select group and was left out", target.name)
				continue
			}
			groups[*target.parentID] = true
		}
		ids = append(ids, target.id)
	}
	return ids
}

func (self *planner) labelNamed(name string, teamID uuid.UUID) *labelTarget {
	var fallback *labelTarget
	for _, target := range self.labelsByName[lower(name)] {
		if target.isGroup {
			continue
		}
		if target.teamID != nil && *target.teamID == teamID {
			return target
		}
		if target.teamID == nil && fallback == nil {
			fallback = target
		}
	}
	return fallback
}
