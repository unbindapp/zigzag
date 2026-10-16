package linear

import (
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/markdown"
	"github.com/unbindapp/zigzag-api/internal/models"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
)

const (
	maxTitleLength = 255
	maxBodyLength  = 100000
)

type issueTarget struct {
	id       uuid.UUID
	team     *teamTarget
	number   int
	renumber bool
	existing bool
	moved    bool
	parentID *uuid.UUID
}

func (self *planner) planIssues() {
	imported := map[string]*ent.Issue{}
	for _, issue := range self.snapshot.Issues {
		if issue.ImportSource == Source && issue.ImportExternalID != "" {
			imported[issue.ImportExternalID] = issue
		}
	}
	withHistory := map[uuid.UUID]bool{}
	for _, id := range self.snapshot.HistoryIssueIDs {
		withHistory[id] = true
	}

	highest := map[*teamTarget]int{}
	claimed := map[*teamTarget]map[int]bool{}
	for _, issue := range self.export.Issues {
		team, ok := self.teams[issue.TeamID]
		if !ok {
			self.warn(EntityIssues, issueRef(issue), "issue belongs to a team that is not imported")
			self.count(EntityIssues, skipped)
			continue
		}
		if found, ok := imported[issue.ID]; ok {
			self.issues[issue.ID] = &issueTarget{id: found.ID, team: team, existing: true, moved: found.TeamID != team.id, parentID: found.ParentID}
			continue
		}
		if claimed[team] == nil {
			claimed[team] = map[int]bool{}
			highest[team] = team.counter
			for number := range team.numbers {
				highest[team] = max(highest[team], number)
			}
		}
		target := &issueTarget{id: uuid.New(), team: team, number: issueNumber(issue)}
		if target.number <= 0 || claimed[team][target.number] || team.numbers[target.number] != uuid.Nil {
			target.renumber = true
		} else {
			claimed[team][target.number] = true
			highest[team] = max(highest[team], target.number)
		}
		self.issues[issue.ID] = target
	}

	for _, issue := range self.export.Issues {
		target, ok := self.issues[issue.ID]
		if !ok {
			continue
		}
		if target.moved {
			self.warn(EntityIssues, issueRef(issue), "issue was moved to another team in Zigzag and was left as is")
			self.count(EntityIssues, skipped)
			continue
		}
		if target.renumber {
			highest[target.team]++
			target.number = highest[target.team]
			self.warn(EntityIssues, issueRef(issue), "number %d is taken in team %s; imported as %s", issueNumber(issue), target.team.key, models.Identifier(target.team.key, target.number))
		}
		if !target.existing && target.number > target.team.counter {
			self.plan.Counters[target.team.id] = max(self.plan.Counters[target.team.id], target.number)
		}
		row := self.issueRow(issue, target)
		self.plan.Issues = append(self.plan.Issues, row)
		if target.existing {
			self.count(EntityIssues, updated)
		} else {
			self.count(EntityIssues, created)
		}
		self.planAttachments(issue, target)
		self.planHistory(issue, target, row, withHistory[target.id])
	}
	inExport := map[string]bool{}
	for _, issue := range self.export.Issues {
		inExport[issue.ID] = true
	}
	for externalID, found := range imported {
		if !inExport[externalID] {
			self.issues[externalID] = &issueTarget{id: found.ID, existing: true, parentID: found.ParentID}
		}
	}
	self.planParents()
}

func issueRef(issue LinearIssue) string {
	if issue.Identifier != "" {
		return issue.Identifier
	}
	return issue.ID
}

func issueNumber(issue LinearIssue) int {
	if issue.Number > 0 {
		return issue.Number
	}
	_, raw, found := strings.Cut(issue.Identifier, "-")
	if !found {
		return 0
	}
	number, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return number
}

func issueTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Untitled"
	}
	return clip(title, maxTitleLength)
}

func (self *planner) issueRow(issue LinearIssue, target *issueTarget) issue_repo.ImportedIssue {
	ref := issueRef(issue)
	description := clip(text(issue.Description), maxBodyLength)
	createdAt := timeOr(issue.CreatedAt, self.options.Now)
	updatedAt := timeOr(issue.UpdatedAt, createdAt)
	row := issue_repo.ImportedIssue{
		ID:                target.id,
		Existing:          target.existing,
		ExternalID:        issue.ID,
		URL:               text(issue.URL),
		TeamID:            target.team.id,
		Number:            target.number,
		Title:             issueTitle(issue.Title),
		Description:       description,
		DescriptionText:   markdown.PlainText(description),
		Priority:          clampPriority(issue.Priority),
		Estimate:          issue.Estimate,
		StateID:           self.issueState(issue, target.team),
		AssigneeID:        self.user(EntityIssues, ref, issue.AssigneeID),
		CreatorID:         self.user(EntityIssues, ref, issue.CreatorID),
		DueDate:           self.date(EntityIssues, ref, "dueDate", issue.DueDate),
		SortOrder:         -float64(target.number),
		SubIssueSortOrder: float64(target.number),
		StartedAt:         utc(issue.StartedAt),
		CompletedAt:       utc(issue.CompletedAt),
		CanceledAt:        utc(issue.CanceledAt),
		TriagedAt:         utc(issue.TriagedAt),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		LabelIDs:          self.issueLabels(issue, target.team.id),
		SubscriberIDs:     []uuid.UUID{},
	}
	if issue.SortOrder != nil {
		row.SortOrder = *issue.SortOrder
	}
	if !self.options.RestoreArchived {
		row.ArchivedAt = utc(issue.ArchivedAt)
	}
	if flag(issue.Trashed) && !self.options.RestoreTrashed {
		row.TrashedAt = &updatedAt
	}
	self.placeIssue(issue, ref, target.team, &row)
	self.appear(target.team.id, row.AssigneeID)
	self.appear(target.team.id, row.CreatorID)
	for _, id := range []*uuid.UUID{row.CreatorID, row.AssigneeID} {
		if id != nil && !slices.Contains(row.SubscriberIDs, *id) {
			row.SubscriberIDs = append(row.SubscriberIDs, *id)
		}
	}
	return row
}

func (self *planner) placeIssue(issue LinearIssue, ref string, team *teamTarget, row *issue_repo.ImportedIssue) {
	if id := text(issue.ProjectID); id != "" {
		if projectID, ok := self.projects[id]; ok {
			row.ProjectID = &projectID
		} else {
			self.warn(EntityIssues, ref, "project %s is not imported; the issue has no project", id)
		}
	}
	if id := text(issue.MilestoneID); id != "" {
		milestone, ok := self.milestones[id]
		switch {
		case !ok:
			self.warn(EntityIssues, ref, "milestone %s is not imported; the issue has no milestone", id)
		case row.ProjectID != nil && *row.ProjectID != milestone.projectID:
			self.warn(EntityIssues, ref, "milestone %s belongs to another project and was left out", id)
		default:
			row.ProjectID = &milestone.projectID
			row.MilestoneID = &milestone.id
		}
	}
	if id := text(issue.CycleID); id != "" {
		cycle, ok := self.cycles[id]
		switch {
		case !ok:
			self.warn(EntityIssues, ref, "cycle %s is not imported; the issue has no cycle", id)
		case cycle.teamID != team.id:
			self.warn(EntityIssues, ref, "cycle %s belongs to another team and was left out", id)
		default:
			row.CycleID = &cycle.id
		}
	}
}

func (self *planner) issueState(issue LinearIssue, team *teamTarget) uuid.UUID {
	if id, ok := self.states[text(issue.StateID)]; ok {
		return id
	}
	if state, ok := self.stateFor(team, issue.StateName, issue.StateType); ok {
		return state.id
	}
	state, _ := team.firstState(schema.StateTypeBacklog, schema.StateTypeUnstarted, schema.StateTypeTriage, schema.StateTypeStarted, schema.StateTypeCompleted, schema.StateTypeCanceled)
	self.warn(EntityIssues, issueRef(issue), "state %q is not in the export; the issue is in %q", text(issue.StateName), state.name)
	return state.id
}

func (self *planner) stateFor(team *teamTarget, name, kind *string) (stateRef, bool) {
	stateKind, known := stateType(text(kind))
	if !known {
		stateKind = ""
	}
	if text(name) != "" {
		if state, ok := team.stateByName(text(name), stateKind); ok {
			return state, true
		}
	}
	if stateKind == "" {
		return stateRef{}, false
	}
	return team.firstState(stateKind)
}

func (self *planner) planParents() {
	for _, issue := range self.export.Issues {
		target, ok := self.issues[issue.ID]
		if !ok || target.moved {
			continue
		}
		var parentID *uuid.UUID
		if id := text(issue.ParentID); id != "" {
			parent, ok := self.issues[id]
			switch {
			case !ok:
				self.warn(EntityIssues, issueRef(issue), "parent issue %s is not imported; the issue is top level", id)
			case parent.id == target.id:
			default:
				parentID = &parent.id
			}
		}
		if !target.existing && parentID == nil {
			continue
		}
		if target.existing && sameID(target.parentID, parentID) {
			continue
		}
		self.plan.Parents = append(self.plan.Parents, issue_repo.ImportedParent{IssueID: target.id, ParentID: parentID})
	}
}

func sameID(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (self *planner) planAttachments(issue LinearIssue, target *issueTarget) {
	existing := map[string]uuid.UUID{}
	if target.existing {
		for _, a := range self.snapshot.Attachments {
			if a.IssueID == target.id {
				existing[a.URL] = a.ID
			}
		}
	}
	for i, a := range issue.Attachments {
		url := strings.TrimSpace(a.URL)
		if url == "" {
			self.warn(EntityAttachments, a.ID, "attachment on %s has no url and was skipped", issueRef(issue))
			self.count(EntityAttachments, skipped)
			continue
		}
		title := text(a.Title)
		if title == "" {
			title = url
		}
		row := issue_repo.ImportedAttachment{
			IssueID:   target.id,
			Title:     title,
			Subtitle:  text(a.Subtitle),
			URL:       url,
			SortOrder: float64(i),
			CreatedAt: timeOr(a.CreatedAt, timeOr(issue.CreatedAt, self.options.Now)),
		}
		if id, ok := existing[url]; ok {
			row.ID = id
			row.Existing = true
			self.count(EntityAttachments, updated)
		} else {
			row.ID = uuid.New()
			existing[url] = row.ID
			self.count(EntityAttachments, created)
		}
		self.plan.Attachments = append(self.plan.Attachments, row)
	}
}

func (self *planner) planHistory(issue LinearIssue, target *issueTarget, row issue_repo.ImportedIssue, recorded bool) {
	imported := schema.FieldChange{Field: "imported", To: row.URL, ToLabel: issueRef(issue)}
	entries := []issue_repo.ImportedHistory{}
	var previous *stateRef
	for _, span := range issue.StateHistory {
		state, ok := self.spanState(span, target.team)
		label := text(span.StateName)
		change := schema.FieldChange{Field: "state", ToLabel: label}
		if ok {
			change.To = state.id.String()
			change.ToLabel = state.name
		}
		if previous != nil {
			change.From = previous.id.String()
			change.FromLabel = previous.name
		}
		changes := []schema.FieldChange{change}
		if len(entries) == 0 {
			changes = []schema.FieldChange{imported, change}
		}
		entries = append(entries, issue_repo.ImportedHistory{IssueID: target.id, CreatedAt: timeOr(span.StartedAt, row.CreatedAt), Changes: changes})
		if ok {
			previous = &state
		} else {
			previous = &stateRef{name: label}
		}
	}
	if len(entries) == 0 {
		entries = append(entries, issue_repo.ImportedHistory{IssueID: target.id, CreatedAt: row.CreatedAt, Changes: []schema.FieldChange{imported}})
	}
	if recorded {
		for range entries {
			self.count(EntityHistory, skipped)
		}
		return
	}
	for range entries {
		self.count(EntityHistory, created)
	}
	self.plan.History = append(self.plan.History, entries...)
}

func (self *planner) spanState(span LinearStateSpan, team *teamTarget) (stateRef, bool) {
	if id, ok := self.states[text(span.StateID)]; ok {
		for _, state := range team.states {
			if state.id == id {
				return state, true
			}
		}
	}
	return self.stateFor(team, span.StateName, span.StateType)
}
