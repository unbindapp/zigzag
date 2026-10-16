package issue_repo

func (suite *IssueRepositorySuite) TestLinkedAttachmentsMatchByURLAndUpdate() {
	issue := suite.create("Limit websocket reads", nil)
	other := suite.create("Other", nil)
	branch, err := suite.repo.CreateAttachment(suite.Ctx, nil, &CreateAttachmentInput{
		IssueID: issue.ID, Title: "unb-1-limit", URL: "https://github.com/acme/api/tree/unb-1-limit", SourceType: "github_branch",
		Metadata: map[string]any{"repo": "acme/api", "branch": "unb-1-limit"},
	})
	suite.Require().NoError(err)
	suite.Equal("acme/api", branch.Metadata["repo"])
	_, err = suite.repo.CreateAttachment(suite.Ctx, nil, &CreateAttachmentInput{IssueID: other.ID, Title: "#7", URL: "https://github.com/acme/api/pull/7", SourceType: "github"})
	suite.Require().NoError(err)

	found, err := suite.repo.AttachmentsByURL(suite.Ctx, nil, issue.ID, []string{"https://github.com/acme/api/pull/7", "https://github.com/acme/api/tree/unb-1-limit"})
	suite.Require().NoError(err)
	suite.Require().Len(found, 1, "attachments of other issues never match")
	suite.Equal(branch.ID, found[0].ID)

	updated, err := suite.repo.UpdateLinkedAttachment(suite.Ctx, nil, branch.ID, &LinkedAttachmentInput{
		Title: "#7 Limit websocket reads", URL: "https://github.com/acme/api/pull/7", SourceType: "github",
		Metadata: map[string]any{"repo": "acme/api", "number": 7, "state": "open"},
	})
	suite.Require().NoError(err)
	suite.Equal("#7 Limit websocket reads", updated.Title)
	suite.Equal("https://github.com/acme/api/pull/7", updated.URL)
	suite.Equal("github", updated.SourceType)
	suite.Equal("open", updated.Metadata["state"])
	suite.Equal(branch.SortOrder, updated.SortOrder)
}
