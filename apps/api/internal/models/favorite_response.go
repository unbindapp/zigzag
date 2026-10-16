package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

type FavoriteResponse struct {
	ID        uuid.UUID           `json:"id" format:"uuid"`
	Kind      schema.FavoriteKind `json:"kind"`
	EntityID  uuid.UUID           `json:"entity_id" format:"uuid"`
	Name      string              `json:"name" doc:"Current name or title of the favorited entity"`
	Icon      string              `json:"icon"`
	Color     string              `json:"color"`
	URLParams map[string]string   `json:"url_params" doc:"Route parameters of the entity's page: identifier for issues, slug for projects, documents and initiatives, key and number for cycles, view_id for views, key for teams, label_id for labels"`
	SortOrder float64             `json:"sort_order"`
	CreatedAt time.Time           `json:"created_at"`
}

func TransformFavoriteEntity(entity *ent.Favorite, name, icon, color string, urlParams map[string]string) *FavoriteResponse {
	return &FavoriteResponse{
		ID:        entity.ID,
		Kind:      entity.Kind,
		EntityID:  entity.EntityID,
		Name:      name,
		Icon:      icon,
		Color:     color,
		URLParams: urlParams,
		SortOrder: entity.SortOrder,
		CreatedAt: entity.CreatedAt,
	}
}
