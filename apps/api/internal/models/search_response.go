package models

type SearchResponse struct {
	Issues    []*IssueSummary     `json:"issues" nullable:"false"`
	Projects  []*ProjectSummary   `json:"projects" nullable:"false"`
	Documents []*DocumentResponse `json:"documents" nullable:"false"`
}
