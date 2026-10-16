package upload_service

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	upload_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/upload"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type UploadServiceSuite struct {
	services.ServiceTestSuite
	service     *UploadService
	userID      uuid.UUID
	workspaceID uuid.UUID
}

func (suite *UploadServiceSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.service = NewUploadService(&config.Config{MaxUploadBytes: 64 << 10}, suite.MockRepo)
	suite.userID = uuid.New()
	suite.workspaceID = uuid.New()
}

func (suite *UploadServiceSuite) allow(role schema.WorkspaceRole, result error) {
	suite.MockPermissionsRepo.EXPECT().Check(suite.Ctx, suite.userID, []permissions_repo.PermissionCheck{{Role: role, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *UploadServiceSuite) expectStored() *upload_repo.CreateUploadInput {
	stored := &upload_repo.CreateUploadInput{}
	suite.MockUploadRepo.EXPECT().Create(suite.Ctx, mock.Anything).RunAndReturn(func(_ context.Context, input *upload_repo.CreateUploadInput) (*ent.Upload, error) {
		*stored = *input
		return &ent.Upload{ID: input.ID, WorkspaceID: input.WorkspaceID, FileName: input.FileName, ContentType: input.ContentType, SizeBytes: int64(len(input.Data)), StorageKey: input.StorageKey, Width: input.Width, Height: input.Height}, nil
	}).Once()
	return stored
}

func encoded(encode func(io.Writer, image.Image) error) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 37, 21))
	img.Set(1, 1, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	_ = encode(&buf, img)
	return buf.Bytes()
}

func webPLossless(width, height int) []byte {
	bits := uint32(width-1) | uint32(height-1)<<14
	chunk := make([]byte, 0, 32)
	chunk = append(chunk, []byte("VP8L")...)
	chunk = binary.LittleEndian.AppendUint32(chunk, 5)
	chunk = append(chunk, 0x2f)
	chunk = binary.LittleEndian.AppendUint32(chunk, bits)
	chunk = append(chunk, make([]byte, 16)...)
	out := append([]byte("RIFF"), binary.LittleEndian.AppendUint32(nil, uint32(4+len(chunk)))...)
	out = append(out, []byte("WEBP")...)
	return append(out, chunk...)
}

func (suite *UploadServiceSuite) TestImagesGetTheirSize() {
	cases := map[string][]byte{
		"png":  encoded(png.Encode),
		"jpeg": encoded(func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) }),
		"gif":  encoded(func(w io.Writer, img image.Image) error { return gif.Encode(w, img, nil) }),
		"webp": webPLossless(37, 21),
	}
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	for name, body := range cases {
		suite.Run(name, func() {
			stored := suite.expectStored()
			resp, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "pic." + name, ContentType: "application/octet-stream", Body: bytes.NewReader(body)})
			suite.Require().NoError(err)
			suite.Equal("image/"+name, resp.ContentType)
			suite.Require().NotNil(resp.Width)
			suite.Equal(37, *resp.Width)
			suite.Equal(21, *resp.Height)
			suite.Equal(body, stored.Data)
			suite.Equal(suite.workspaceID.String()+"/"+stored.ID.String()+"."+name, stored.StorageKey, "files live under the workspace id")
			suite.Equal("/api/go/uploads/get?upload_id="+stored.ID.String(), resp.URL)
		})
	}
}

func (suite *UploadServiceSuite) TestOtherFilesKeepTheirDeclaredType() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	stored := suite.expectStored()
	resp, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "../../notes.TXT", ContentType: "text/plain; charset=utf-8", Body: bytes.NewReader([]byte("hello"))})
	suite.Require().NoError(err)
	suite.Equal("text/plain", resp.ContentType)
	suite.Nil(resp.Width)
	suite.Equal("notes.TXT", stored.FileName, "path segments are dropped from file names")
}

func (suite *UploadServiceSuite) TestSVGsAreValidatedAndSized() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	cases := []struct {
		name     string
		declared string
		body     string
		width    *int
		height   *int
	}{
		{name: "declared svg with viewBox", declared: "image/svg+xml; charset=utf-8", body: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 16"><path d="M0 0h24v16z"/></svg>`, width: utils.ToPtr(24), height: utils.ToPtr(16)},
		{name: "sniffed svg without size", declared: "application/octet-stream", body: `<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><circle r="4"/></svg>`},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			stored := suite.expectStored()
			resp, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "logo.svg", ContentType: tc.declared, Body: bytes.NewReader([]byte(tc.body))})
			suite.Require().NoError(err)
			suite.Equal("image/svg+xml", resp.ContentType)
			suite.Equal(tc.width, resp.Width)
			suite.Equal(tc.height, resp.Height)
			suite.Equal(int64(len(tc.body)), resp.SizeBytes)
			suite.Equal(tc.body, string(stored.Data))
		})
	}
}

func (suite *UploadServiceSuite) TestUnsafeSVGsAreRefused() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	cases := map[string]struct {
		declared string
		body     string
	}{
		"declared svg with a script":     {declared: "image/svg+xml", body: `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`},
		"sniffed svg with onload":        {declared: "text/plain", body: `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>`},
		"declared svg that is html":      {declared: "image/svg+xml", body: `<html><body>hi</body></html>`},
		"declared svg larger than 1 MiB": {declared: "image/svg+xml", body: `<svg xmlns="http://www.w3.org/2000/svg">` + strings.Repeat(" ", 1<<20) + `</svg>`},
	}
	for name, tc := range cases {
		suite.Run(name, func() {
			_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "logo.svg", ContentType: tc.declared, Body: bytes.NewReader([]byte(tc.body))})
			suite.ErrorIs(err, errdefs.ErrInvalidInput)
		})
	}
}

func (suite *UploadServiceSuite) TestTooLargeFilesAreRefused() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "big.bin", Body: bytes.NewReader(make([]byte, 65<<10))})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *UploadServiceSuite) TestEmptyFilesAreRefused() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, nil)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "empty", Body: bytes.NewReader(nil)})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
}

func (suite *UploadServiceSuite) TestViewersCannotUpload() {
	suite.allow(schema.WorkspaceRoleViewer, nil)
	suite.allow(schema.WorkspaceRoleMember, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "x", Body: bytes.NewReader([]byte("x"))})
	suite.ErrorIs(err, errdefs.ErrUnauthorized, "a viewer sees the workspace, so the refusal is forbidden, not not found")
}

func (suite *UploadServiceSuite) TestOutsidersDoNotLearnTheWorkspaceExists() {
	suite.allow(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.Create(suite.Ctx, suite.userID, &CreateUploadInput{WorkspaceID: suite.workspaceID, FileName: "x", Body: bytes.NewReader([]byte("x"))})
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func (suite *UploadServiceSuite) TestOpenChecksTheWorkspaceAndMarksSafeTypesInline() {
	html := &ent.Upload{ID: uuid.New(), WorkspaceID: suite.workspaceID, ContentType: "text/html"}
	pngUpload := &ent.Upload{ID: uuid.New(), WorkspaceID: suite.workspaceID, ContentType: "image/png"}
	suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, html.ID).Return(html, nil)
	suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, pngUpload.ID).Return(pngUpload, nil)
	suite.MockUploadRepo.EXPECT().GetData(suite.Ctx, html.ID).Return([]byte("<b>hi</b>"), nil)
	suite.MockUploadRepo.EXPECT().GetData(suite.Ctx, pngUpload.ID).Return([]byte("png"), nil)

	suite.allow(schema.WorkspaceRoleViewer, nil)
	opened, err := suite.service.Open(suite.Ctx, suite.userID, html.ID)
	suite.Require().NoError(err)
	suite.Equal("<b>hi</b>", string(opened.Data))
	suite.False(opened.Inline, "html is never rendered inline")

	opened, err = suite.service.Open(suite.Ctx, suite.userID, pngUpload.ID)
	suite.Require().NoError(err)
	suite.True(opened.Inline)
}

func (suite *UploadServiceSuite) TestOpenServesOnlySafeSVGsInline() {
	cases := map[string]struct {
		body   string
		inline bool
	}{
		"clean svg":            {body: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 4"/>`, inline: true},
		"svg stored unchecked": {body: `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`, inline: false},
	}
	suite.allow(schema.WorkspaceRoleViewer, nil)
	for name, tc := range cases {
		suite.Run(name, func() {
			upload := &ent.Upload{ID: uuid.New(), WorkspaceID: suite.workspaceID, ContentType: "image/svg+xml"}
			suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, upload.ID).Return(upload, nil).Once()
			suite.MockUploadRepo.EXPECT().GetData(suite.Ctx, upload.ID).Return([]byte(tc.body), nil).Once()
			opened, err := suite.service.Open(suite.Ctx, suite.userID, upload.ID)
			suite.Require().NoError(err)
			suite.Equal(tc.body, string(opened.Data))
			suite.Equal(tc.inline, opened.Inline)
		})
	}
}

func (suite *UploadServiceSuite) TestOpenMasksOtherWorkspaces() {
	upload := &ent.Upload{ID: uuid.New(), WorkspaceID: suite.workspaceID}
	suite.MockUploadRepo.EXPECT().GetByID(suite.Ctx, upload.ID).Return(upload, nil).Once()
	suite.allow(schema.WorkspaceRoleViewer, errdefs.ErrUnauthorized)
	_, err := suite.service.Open(suite.Ctx, suite.userID, upload.ID)
	suite.ErrorIs(err, errdefs.ErrNotFound)
}

func TestUploadServiceSuite(t *testing.T) {
	suite.Run(t, new(UploadServiceSuite))
}
