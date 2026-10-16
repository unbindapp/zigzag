package linear

import (
	"strings"

	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/names"
	team_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/team"
	user_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/user"
	workspace_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/workspace"
)

const (
	mappingMatched = "matched"
	mappingCreated = "created"
	mappingSkipped = "skipped"
)

func IsBot(email string) bool {
	return strings.HasSuffix(lower(email), "@linear.linear.app")
}

func DisplayNameBases(export *LinearExport) []string {
	bases := make([]string, 0, len(export.Users))
	for _, u := range export.Users {
		if IsBot(u.Email) {
			continue
		}
		bases = append(bases, displayNameBase(u))
	}
	return bases
}

func displayNameBase(u LinearUser) string {
	if handle, err := names.DisplayName(text(u.DisplayName)); err == nil {
		return handle
	}
	return names.SuggestDisplayName(u.Email)
}

func (self *planner) planUsers() {
	existing := map[string]uuid.UUID{}
	for _, u := range self.snapshot.Users {
		existing[lower(u.Email)] = u.ID
	}
	members := map[uuid.UUID]bool{}
	for _, m := range self.snapshot.Members {
		members[m.UserID] = true
	}
	taken := append([]string{}, self.snapshot.TakenDisplayNames...)

	for _, u := range self.export.Users {
		email := lower(u.Email)
		mapping := schemaMapping(u.ID, email)
		var userID *uuid.UUID
		switch {
		case IsBot(email):
			mapping.Action = mappingSkipped
		case existing[email] != uuid.Nil:
			id := existing[email]
			userID = &id
			mapping.Action = mappingMatched
		case !self.options.CreateMissingUsers:
			mapping.Action = mappingSkipped
		default:
			normalized, err := names.Email(email)
			if err != nil {
				self.warn(EntityUsers, u.ID, "email %q is not valid; the user was skipped", u.Email)
				mapping.Action = mappingSkipped
				break
			}
			handle := names.UniqueSeeded(displayNameBase(u), taken, names.MaxLength, u.ID)
			taken = append(taken, handle)
			id := uuid.New()
			userID = &id
			mapping.Action = mappingCreated
			self.plan.Users = append(self.plan.Users, user_repo.ImportedUser{ID: id, Email: normalized, Name: userName(u), DisplayName: handle})
			existing[normalized] = id
		}

		mapping.UserID = userID
		self.plan.UserMapping = append(self.plan.UserMapping, mapping)
		self.users[u.ID] = userID
		switch mapping.Action {
		case mappingCreated:
			self.count(EntityUsers, created)
		case mappingMatched:
			self.count(EntityUsers, updated)
		default:
			self.count(EntityUsers, skipped)
		}

		if userID == nil || members[*userID] {
			continue
		}
		members[*userID] = true
		role := schema.WorkspaceRoleMember
		if u.IsAdmin {
			role = schema.WorkspaceRoleAdmin
		}
		self.plan.Members = append(self.plan.Members, workspace_repo.ImportedMember{UserID: *userID, Role: role})
	}
}

func schemaMapping(linearID, email string) schema.ImportUserMapping {
	return schema.ImportUserMapping{LinearID: linearID, Email: email}
}

func userName(u LinearUser) string {
	for _, candidate := range []string{text(u.Name), text(u.DisplayName)} {
		if candidate != "" {
			return candidate
		}
	}
	local, _, _ := strings.Cut(u.Email, "@")
	return local
}

func (self *planner) user(entity, externalID string, linearID *string) *uuid.UUID {
	id := text(linearID)
	if id == "" {
		return nil
	}
	userID, known := self.users[id]
	if known {
		return userID
	}
	if !self.unknownUsers[id] {
		self.unknownUsers[id] = true
		self.warn(entity, externalID, "user %s is not in the export's users and was left out", id)
	}
	return nil
}

func (self *planner) planTeamMembers() {
	existing := map[[2]uuid.UUID]bool{}
	for _, m := range self.snapshot.TeamMembers {
		existing[[2]uuid.UUID{m.TeamID, m.UserID}] = true
	}
	for _, target := range self.orderedTeams() {
		for userID := range self.appearances[target.id] {
			key := [2]uuid.UUID{target.id, userID}
			if existing[key] {
				continue
			}
			existing[key] = true
			self.plan.TeamMembers = append(self.plan.TeamMembers, team_repo.ImportedTeamMember{TeamID: target.id, UserID: userID})
		}
	}
}
