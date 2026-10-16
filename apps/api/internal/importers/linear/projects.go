package linear

import (
	"cmp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/utils"
	cycle_repo "github.com/unbindapp/zigzag-api/internal/repositories/cycle"
	project_repo "github.com/unbindapp/zigzag-api/internal/repositories/project"
)

const (
	slugBaseLength       = 48
	maxProjectNameLength = 80
)

type milestoneTarget struct {
	id        uuid.UUID
	projectID uuid.UUID
}

type cycleTarget struct {
	id     uuid.UUID
	teamID uuid.UUID
}

func projectSlug(name, linearID string) string {
	base := utils.TruncateSlug(utils.Slugify(name), slugBaseLength)
	if base == "" {
		base = "project"
	}
	suffix := strings.ReplaceAll(linearID, "-", "")
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return base + "-" + strings.ToLower(suffix)
}

func (self *planner) projectStatus(status LinearProjectStatus) *ent.ProjectStatus {
	statuses := slices.Clone(self.snapshot.ProjectStatuses)
	slices.SortStableFunc(statuses, func(a, b *ent.ProjectStatus) int { return cmp.Compare(a.Position, b.Position) })
	var ofType *ent.ProjectStatus
	for _, candidate := range statuses {
		if string(candidate.Type) != lower(status.Type) {
			continue
		}
		if lower(candidate.Name) == lower(status.Name) {
			return candidate
		}
		if ofType == nil {
			ofType = candidate
		}
	}
	if ofType != nil {
		return ofType
	}
	for _, candidate := range statuses {
		if candidate.Type == schema.ProjectStatusBacklog {
			return candidate
		}
	}
	if len(statuses) > 0 {
		return statuses[0]
	}
	return nil
}

func (self *planner) planProjects() {
	existing := map[string]*ent.Project{}
	for _, p := range self.snapshot.Projects {
		existing[p.ImportExternalID] = p
	}

	for _, p := range self.export.Projects {
		status := self.projectStatus(p.Status)
		if status == nil {
			self.warn(EntityProjects, p.ID, "the workspace has no project statuses; project %q was skipped", p.Name)
			self.count(EntityProjects, skipped)
			continue
		}
		if p.Status.Type != "" && string(status.Type) != lower(p.Status.Type) {
			self.warn(EntityProjects, p.ID, "no project status of type %q; project %q uses %q", p.Status.Type, p.Name, status.Name)
		}
		createdAt := timeOr(p.CreatedAt, self.options.Now)
		updatedAt := timeOr(p.UpdatedAt, createdAt)
		row := project_repo.ImportedProject{
			ExternalID:  p.ID,
			Name:        clip(strings.TrimSpace(p.Name), maxProjectNameLength),
			Summary:     text(p.Summary),
			Description: text(p.Description),
			StatusID:    status.ID,
			LeadID:      self.user(EntityProjects, p.ID, p.LeadID),
			Priority:    clampPriority(p.Priority),
			StartDate:   self.date(EntityProjects, p.ID, "startDate", p.StartDate),
			TargetDate:  self.date(EntityProjects, p.ID, "targetDate", p.TargetDate),
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			StartedAt:   utc(p.StartedAt),
			CompletedAt: utc(p.CompletedAt),
			CanceledAt:  utc(p.CanceledAt),
			TeamIDs:     []uuid.UUID{},
			MemberIDs:   []uuid.UUID{},
			LabelIDs:    []uuid.UUID{},
		}
		if row.Name == "" {
			row.Name = "Untitled project"
		}
		if p.SortOrder != nil {
			row.SortOrder = *p.SortOrder
		}
		if !self.options.RestoreArchived {
			row.ArchivedAt = utc(p.ArchivedAt)
		}
		if flag(p.Trashed) && !self.options.RestoreTrashed {
			row.TrashedAt = &updatedAt
		}
		for _, id := range p.TeamIDs {
			if team, ok := self.teams[id]; ok && !slices.Contains(row.TeamIDs, team.id) {
				row.TeamIDs = append(row.TeamIDs, team.id)
			}
		}
		for _, id := range p.MemberIDs {
			if member := self.user(EntityProjects, p.ID, &id); member != nil && !slices.Contains(row.MemberIDs, *member) {
				row.MemberIDs = append(row.MemberIDs, *member)
			}
		}
		for _, id := range p.LabelIDs {
			label, ok := self.labels[id]
			if !ok || label.teamID != nil || label.isGroup {
				self.warn(EntityProjects, p.ID, "label %s is not a workspace label and was left off the project", id)
				continue
			}
			row.LabelIDs = append(row.LabelIDs, label.id)
		}

		if found, ok := existing[p.ID]; ok {
			row.ID = found.ID
			row.Existing = true
			row.Slug = found.Slug
			self.count(EntityProjects, updated)
		} else {
			row.ID = uuid.New()
			row.Slug = projectSlug(row.Name, p.ID)
			self.count(EntityProjects, created)
		}
		self.projects[p.ID] = row.ID
		self.plan.Projects = append(self.plan.Projects, row)
	}
	for _, p := range self.snapshot.Projects {
		if _, ok := self.projects[p.ImportExternalID]; !ok {
			self.projects[p.ImportExternalID] = p.ID
		}
	}
	self.planMilestones()
}

func (self *planner) planMilestones() {
	existing := map[uuid.UUID]map[string]uuid.UUID{}
	for _, m := range self.snapshot.Milestones {
		if existing[m.ProjectID] == nil {
			existing[m.ProjectID] = map[string]uuid.UUID{}
		}
		existing[m.ProjectID][lower(m.Name)] = m.ID
	}
	for i, m := range self.export.Milestones {
		projectID, ok := self.projects[m.ProjectID]
		if !ok {
			self.warn(EntityMilestones, m.ID, "milestone %q belongs to a project that is not imported", m.Name)
			self.count(EntityMilestones, skipped)
			continue
		}
		name := clip(strings.TrimSpace(m.Name), maxProjectNameLength)
		if name == "" {
			name = "Untitled milestone"
		}
		row := project_repo.ImportedMilestone{
			ProjectID:   projectID,
			Name:        name,
			Description: text(m.Description),
			TargetDate:  self.date(EntityMilestones, m.ID, "targetDate", m.TargetDate),
			SortOrder:   float64(i),
		}
		if m.SortOrder != nil {
			row.SortOrder = *m.SortOrder
		}
		if id, ok := existing[projectID][lower(name)]; ok {
			row.ID = id
			row.Existing = true
			self.count(EntityMilestones, updated)
		} else {
			row.ID = uuid.New()
			if existing[projectID] == nil {
				existing[projectID] = map[string]uuid.UUID{}
			}
			existing[projectID][lower(name)] = row.ID
			self.count(EntityMilestones, created)
		}
		self.plan.Milestones = append(self.plan.Milestones, row)
		self.milestones[m.ID] = milestoneTarget{id: row.ID, projectID: projectID}
	}
}

func (self *planner) planCycles() {
	existing := map[cycleKey]uuid.UUID{}
	for _, c := range self.snapshot.Cycles {
		existing[cycleKey{c.TeamID, c.Number}] = c.ID
	}
	for _, c := range self.export.Cycles {
		team, ok := self.teams[c.TeamID]
		if !ok {
			self.warn(EntityCycles, c.ID, "cycle %d belongs to a team that is not imported", c.Number)
			self.count(EntityCycles, skipped)
			continue
		}
		key := cycleKey{team.id, c.Number}
		if id, ok := existing[key]; ok {
			self.cycles[c.ID] = cycleTarget{id: id, teamID: team.id}
			self.count(EntityCycles, skipped)
			continue
		}
		if c.StartsAt == nil || c.EndsAt == nil || !c.EndsAt.After(*c.StartsAt) {
			self.warn(EntityCycles, c.ID, "cycle %d has no valid start and end and was skipped", c.Number)
			self.count(EntityCycles, skipped)
			continue
		}
		row := cycle_repo.ImportedCycle{
			ID:          uuid.New(),
			TeamID:      team.id,
			Number:      c.Number,
			Name:        text(c.Name),
			Description: text(c.Description),
			StartsAt:    c.StartsAt.UTC(),
			EndsAt:      c.EndsAt.UTC(),
			CompletedAt: utc(c.CompletedAt),
		}
		existing[key] = row.ID
		self.cycles[c.ID] = cycleTarget{id: row.ID, teamID: team.id}
		self.plan.Cycles = append(self.plan.Cycles, row)
		self.count(EntityCycles, created)
	}
}

type cycleKey struct {
	teamID uuid.UUID
	number int
}

func clampPriority(priority int) int {
	return min(max(priority, 0), 4)
}
