package models

type PaginationInput struct {
	Limit  int    `query:"limit" minimum:"1" maximum:"250" default:"50" doc:"Page size, 1 to 250"`
	Cursor string `query:"cursor" required:"false" doc:"Opaque cursor from the previous page's next_cursor"`
}
