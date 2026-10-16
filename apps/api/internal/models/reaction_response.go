package models

import (
	"github.com/google/uuid"
	"github.com/zigzagapp-org/zigzag-api/ent"
)

type ReactionGroup struct {
	Emoji   string         `json:"emoji"`
	Count   int            `json:"count"`
	Users   []*UserSummary `json:"users" nullable:"false"`
	Reacted bool           `json:"reacted" doc:"Whether the caller added this emoji"`
}

func TransformReactionGroups(entities []*ent.Reaction, callerID uuid.UUID) []*ReactionGroup {
	groups := []*ReactionGroup{}
	byEmoji := map[string]*ReactionGroup{}
	for _, entity := range entities {
		group, ok := byEmoji[entity.Emoji]
		if !ok {
			group = &ReactionGroup{Emoji: entity.Emoji, Users: []*UserSummary{}}
			byEmoji[entity.Emoji] = group
			groups = append(groups, group)
		}
		group.Count++
		if entity.UserID == callerID {
			group.Reacted = true
		}
		if user := TransformUserSummary(entity.Edges.User); user != nil {
			group.Users = append(group.Users, user)
		}
	}
	return groups
}
