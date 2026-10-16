package upload_service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/config"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/internal/infrastructure/storage"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
	"github.com/zigzagapp-org/zigzag-api/internal/services"
)

type memObject struct {
	body        []byte
	contentType string
}

type memStore struct {
	objects map[string]memObject
	putErr  error
}

func newMemStore() *memStore {
	return &memStore{objects: map[string]memObject{}}
}

func (self *memStore) Put(ctx context.Context, key string, reader io.Reader, contentType string) error {
	if self.putErr != nil {
		return self.putErr
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	self.objects[key] = memObject{body: body, contentType: contentType}
	return nil
}

func (self *memStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	object, ok := self.objects[key]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(object.body)), nil
}

func (self *memStore) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := self.objects[key]
	return ok, nil
}

func (self *memStore) Delete(ctx context.Context, key string) error {
	delete(self.objects, key)
	return nil
}

type MigrateSuite struct {
	services.ServiceTestSuite
	local   *memStore
	remote  *memStore
	service *UploadService
}

func (suite *MigrateSuite) SetupTest() {
	suite.ServiceTestSuite.SetupTest()
	suite.local = newMemStore()
	suite.remote = newMemStore()
	suite.service = NewUploadService(&config.Config{}, suite.MockRepo, suite.remote)
}

func (suite *MigrateSuite) upload(key, contentType string) *ent.Upload {
	return &ent.Upload{ID: uuid.New(), StorageKey: key, ContentType: contentType}
}

func (suite *MigrateSuite) listReturns(after uuid.UUID, page []*ent.Upload) {
	suite.MockUploadRepo.EXPECT().ListAfter(mock.Anything, after, copyPageSize).Return(page, nil).Once()
}

func (suite *MigrateSuite) TestCopiesWhatTheBucketLacksAndKeepsLocalFiles() {
	onlyLocal := suite.upload("ws/a.png", "image/png")
	inBoth := suite.upload("ws/b.txt", "text/plain")
	lost := suite.upload("ws/c.pdf", "application/pdf")
	suite.local.objects[onlyLocal.StorageKey] = memObject{body: []byte("png")}
	suite.local.objects[inBoth.StorageKey] = memObject{body: []byte("local")}
	suite.remote.objects[inBoth.StorageKey] = memObject{body: []byte("remote"), contentType: "text/plain"}

	suite.listReturns(uuid.Nil, []*ent.Upload{onlyLocal, inBoth, lost})
	result, err := suite.service.CopyFrom(suite.Ctx, suite.local)
	suite.Require().NoError(err)
	suite.Equal(&CopyResult{Copied: 1, Missing: 1}, result)
	suite.Equal(memObject{body: []byte("png"), contentType: "image/png"}, suite.remote.objects[onlyLocal.StorageKey], "the stored content type goes with the object")
	suite.Equal([]byte("remote"), suite.remote.objects[inBoth.StorageKey].body, "objects already in the bucket are left alone")
	suite.Len(suite.local.objects, 2, "local files are never deleted")

	suite.listReturns(uuid.Nil, []*ent.Upload{onlyLocal, inBoth, lost})
	result, err = suite.service.CopyFrom(suite.Ctx, suite.local)
	suite.Require().NoError(err)
	suite.Equal(&CopyResult{Missing: 1}, result, "a second run copies nothing")
}

func (suite *MigrateSuite) TestPagesThroughEveryUpload() {
	first := make([]*ent.Upload, copyPageSize)
	for i := range first {
		first[i] = suite.upload("ws/"+uuid.NewString(), "image/png")
		suite.local.objects[first[i].StorageKey] = memObject{body: []byte("x")}
	}
	last := suite.upload("ws/last", "image/png")
	suite.local.objects[last.StorageKey] = memObject{body: []byte("x")}

	suite.listReturns(uuid.Nil, first)
	suite.listReturns(first[len(first)-1].ID, []*ent.Upload{last})
	result, err := suite.service.CopyFrom(suite.Ctx, suite.local)
	suite.Require().NoError(err)
	suite.Equal(copyPageSize+1, result.Copied)
}

func (suite *MigrateSuite) TestFailedCopiesAreCountedAndTheRestContinue() {
	a := suite.upload("ws/a", "image/png")
	b := suite.upload("ws/b", "image/png")
	suite.local.objects[a.StorageKey] = memObject{body: []byte("a")}
	suite.local.objects[b.StorageKey] = memObject{body: []byte("b")}
	suite.remote.putErr = errors.New("bucket unavailable")

	suite.listReturns(uuid.Nil, []*ent.Upload{a, b})
	result, err := suite.service.CopyFrom(suite.Ctx, suite.local)
	suite.Require().NoError(err)
	suite.Equal(&CopyResult{Failed: 2}, result)
}

func (suite *MigrateSuite) TestListErrorStopsTheCopy() {
	suite.MockUploadRepo.EXPECT().ListAfter(mock.Anything, uuid.Nil, copyPageSize).Return(nil, errors.New("db down")).Once()
	_, err := suite.service.CopyFrom(suite.Ctx, suite.local)
	suite.Error(err)
}

func (suite *MigrateSuite) TestMigrateRunsAsTheSystemCaller() {
	suite.MockUploadRepo.EXPECT().ListAfter(mock.MatchedBy(permissions_repo.IsSystemCaller), uuid.Nil, copyPageSize).Return(nil, nil).Once()
	suite.service.MigrateFrom(suite.Ctx, suite.local)
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(MigrateSuite))
}
