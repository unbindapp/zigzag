package storage

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ StoreInterface = (*S3Store)(nil)

type fakeObject struct {
	body        string
	contentType string
}

type fakeS3 struct {
	t           *testing.T
	bucket      string
	accessKeyID string
	failPart    int

	mu        sync.Mutex
	objects   map[string]fakeObject
	uploads   map[string]map[int]string
	uploadCT  map[string]string
	aborted   []string
	completed int
}

func newFakeS3(t *testing.T) (*fakeS3, *httptest.Server) {
	fake := &fakeS3{t: t, bucket: "zigzag", accessKeyID: "zz-key", objects: map[string]fakeObject{}, uploads: map[string]map[int]string{}, uploadCT: map[string]string{}}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	return fake, server
}

func (self *fakeS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Authorization"), "Credential="+self.accessKeyID+"/") {
		self.fail(w, http.StatusForbidden, "InvalidAccessKeyId")
		return
	}
	bucket, key, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if bucket != self.bucket {
		self.fail(w, http.StatusNotFound, "NoSuchBucket")
		return
	}
	self.mu.Lock()
	defer self.mu.Unlock()
	query := r.URL.Query()
	switch {
	case key == "" && r.Method == http.MethodHead:
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodPost && query.Has("uploads"):
		self.createMultipart(w, r, key)
	case r.Method == http.MethodPut && query.Has("partNumber"):
		self.uploadPart(w, r, query.Get("uploadId"), query.Get("partNumber"))
	case r.Method == http.MethodPost && query.Has("uploadId"):
		self.completeMultipart(w, r, key, query.Get("uploadId"))
	case r.Method == http.MethodDelete && query.Has("uploadId"):
		self.aborted = append(self.aborted, query.Get("uploadId"))
		delete(self.uploads, query.Get("uploadId"))
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		self.objects[key] = fakeObject{body: string(body), contentType: r.Header.Get("Content-Type")}
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodGet || r.Method == http.MethodHead:
		self.getObject(w, r, key)
	case r.Method == http.MethodDelete:
		delete(self.objects, key)
		w.WriteHeader(http.StatusNoContent)
	default:
		self.fail(w, http.StatusNotImplemented, "NotImplemented")
	}
}

func (self *fakeS3) getObject(w http.ResponseWriter, r *http.Request, key string) {
	object, ok := self.objects[key]
	if !ok {
		self.fail(w, http.StatusNotFound, "NoSuchKey")
		return
	}
	w.Header().Set("Content-Type", object.contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(object.body)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = io.WriteString(w, object.body)
	}
}

func (self *fakeS3) createMultipart(w http.ResponseWriter, r *http.Request, key string) {
	uploadID := fmt.Sprintf("upload-%d", len(self.uploads)+len(self.aborted)+self.completed+1)
	self.uploads[uploadID] = map[int]string{}
	self.uploadCT[uploadID] = r.Header.Get("Content-Type")
	_, _ = fmt.Fprintf(w, `<InitiateMultipartUploadResult><Bucket>%s</Bucket><Key>%s</Key><UploadId>%s</UploadId></InitiateMultipartUploadResult>`, self.bucket, key, uploadID)
}

func (self *fakeS3) uploadPart(w http.ResponseWriter, r *http.Request, uploadID, partNumber string) {
	number, _ := strconv.Atoi(partNumber)
	parts, ok := self.uploads[uploadID]
	if !ok {
		self.fail(w, http.StatusNotFound, "NoSuchUpload")
		return
	}
	if number == self.failPart {
		self.fail(w, http.StatusBadRequest, "InvalidPart")
		return
	}
	body, _ := io.ReadAll(r.Body)
	parts[number] = string(body)
	w.Header().Set("ETag", fmt.Sprintf(`"etag-%d"`, number))
	w.WriteHeader(http.StatusOK)
}

func (self *fakeS3) completeMultipart(w http.ResponseWriter, r *http.Request, key, uploadID string) {
	var request struct {
		Parts []struct {
			PartNumber int
			ETag       string
		} `xml:"Part"`
	}
	require.NoError(self.t, xml.NewDecoder(r.Body).Decode(&request))
	parts := self.uploads[uploadID]
	numbers := []int{}
	for _, part := range request.Parts {
		assert.Equal(self.t, fmt.Sprintf(`"etag-%d"`, part.PartNumber), part.ETag)
		numbers = append(numbers, part.PartNumber)
	}
	assert.True(self.t, sort.IntsAreSorted(numbers))
	var body strings.Builder
	for _, number := range numbers {
		body.WriteString(parts[number])
	}
	self.objects[key] = fakeObject{body: body.String(), contentType: self.uploadCT[uploadID]}
	delete(self.uploads, uploadID)
	self.completed++
	_, _ = fmt.Fprintf(w, `<CompleteMultipartUploadResult><Bucket>%s</Bucket><Key>%s</Key><ETag>"done"</ETag></CompleteMultipartUploadResult>`, self.bucket, key)
}

func (self *fakeS3) fail(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `<Error><Code>%s</Code><Message>%s</Message></Error>`, code, code)
}

func newTestS3Store(t *testing.T, endpoint, bucket string) *S3Store {
	store, err := NewS3Store(context.Background(), S3Options{
		Endpoint:        endpoint,
		Bucket:          bucket,
		AccessKeyID:     "zz-key",
		SecretAccessKey: "zz-secret",
		Region:          "us-east-1",
		UsePathStyle:    true,
	})
	require.NoError(t, err)
	return store
}

func TestS3CheckBucket(t *testing.T) {
	_, server := newFakeS3(t)
	ctx := context.Background()

	assert.NoError(t, newTestS3Store(t, server.URL, "zigzag").CheckBucket(ctx))
	assert.Error(t, newTestS3Store(t, server.URL, "missing").CheckBucket(ctx))
}

func TestS3PutOpenExistsDelete(t *testing.T) {
	fake, server := newFakeS3(t)
	store := newTestS3Store(t, server.URL, "zigzag")
	ctx := context.Background()

	require.NoError(t, store.Put(ctx, "ws-1/file.png", strings.NewReader("hello"), "image/png"))
	assert.Equal(t, fakeObject{body: "hello", contentType: "image/png"}, fake.objects["uploads/ws-1/file.png"], "objects keep the uploads/<workspace>/<file> layout")

	body, err := store.Open(ctx, "ws-1/file.png")
	require.NoError(t, err)
	data, _ := io.ReadAll(body)
	body.Close()
	assert.Equal(t, "hello", string(data))

	exists, err := store.Exists(ctx, "ws-1/file.png")
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, store.Delete(ctx, "ws-1/file.png"))
	exists, err = store.Exists(ctx, "ws-1/file.png")
	require.NoError(t, err)
	assert.False(t, exists)
	_, err = store.Open(ctx, "ws-1/file.png")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.NoError(t, store.Delete(ctx, "ws-1/file.png"), "deleting a missing object is not an error")
}

func TestS3PutStreamsLargeBodiesInParts(t *testing.T) {
	fake, server := newFakeS3(t)
	store := newTestS3Store(t, server.URL, "zigzag")
	store.partSize = 4
	ctx := context.Background()

	for name, body := range map[string]string{"uneven": "0123456789", "exact parts": "01234567"} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, store.Put(ctx, "ws/"+strings.ReplaceAll(name, " ", "-"), strings.NewReader(body), "video/mp4"))
			assert.Equal(t, fakeObject{body: body, contentType: "video/mp4"}, fake.objects["uploads/ws/"+strings.ReplaceAll(name, " ", "-")])
		})
	}
	assert.Equal(t, 2, fake.completed)
	assert.Empty(t, fake.uploads)
}

func TestS3FailedPartAbortsTheUpload(t *testing.T) {
	fake, server := newFakeS3(t)
	fake.failPart = 2
	store := newTestS3Store(t, server.URL, "zigzag")
	store.partSize = 4

	assert.Error(t, store.Put(context.Background(), "ws/file", strings.NewReader("0123456789"), "video/mp4"))
	assert.Len(t, fake.aborted, 1)
	assert.NotContains(t, fake.objects, "uploads/ws/file")
}

func TestS3KeyPrefixIsAppliedOnce(t *testing.T) {
	fake, server := newFakeS3(t)
	ctx := context.Background()

	for _, prefix := range []string{"dux", "dux/", "/dux/"} {
		t.Run(prefix, func(t *testing.T) {
			store, err := NewS3Store(ctx, S3Options{Endpoint: server.URL, Bucket: "zigzag", AccessKeyID: "zz-key", SecretAccessKey: "zz-secret", Region: "auto", UsePathStyle: true, KeyPrefix: prefix})
			require.NoError(t, err)
			require.NoError(t, store.Put(ctx, "ws/file.png", strings.NewReader(prefix), "image/png"))
			assert.Equal(t, prefix, fake.objects["dux/uploads/ws/file.png"].body)

			body, err := store.Open(ctx, "ws/file.png")
			require.NoError(t, err)
			data, _ := io.ReadAll(body)
			body.Close()
			assert.Equal(t, prefix, string(data))
		})
	}
	assert.Len(t, fake.objects, 1)

	unprefixed := newTestS3Store(t, server.URL, "zigzag")
	exists, err := unprefixed.Exists(ctx, "ws/file.png")
	require.NoError(t, err)
	assert.False(t, exists, "instances with different prefixes do not see each other's objects")
}

func TestS3KeysAreValidated(t *testing.T) {
	_, server := newFakeS3(t)
	store := newTestS3Store(t, server.URL, "zigzag")
	ctx := context.Background()

	for _, key := range []string{"", "../escape", "/etc/passwd", "a/../b"} {
		assert.ErrorIs(t, store.Put(ctx, key, strings.NewReader("x"), ""), ErrInvalidKey, key)
		_, err := store.Open(ctx, key)
		assert.ErrorIs(t, err, ErrInvalidKey, key)
	}
}
