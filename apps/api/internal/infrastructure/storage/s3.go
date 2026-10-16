package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const defaultPartSize = 8 << 20

type S3Options struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UsePathStyle    bool
	KeyPrefix       string
}

type S3Store struct {
	client    *s3.Client
	bucket    string
	keyPrefix string
	partSize  int
}

func NewS3Store(ctx context.Context, options S3Options) (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(options.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(options.AccessKeyID, options.SecretAccessKey, "")),
		// R2 rejects the default "when supported" checksum behaviour
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if options.Endpoint != "" {
			o.BaseEndpoint = aws.String(options.Endpoint)
		}
		o.UsePathStyle = options.UsePathStyle
	})
	return &S3Store{client: client, bucket: options.Bucket, keyPrefix: normalizePrefix(options.KeyPrefix), partSize: defaultPartSize}, nil
}

func (self *S3Store) CheckBucket(ctx context.Context) error {
	_, err := self.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(self.bucket)})
	return err
}

func (self *S3Store) Put(ctx context.Context, key string, reader io.Reader, contentType string) error {
	objectKey, err := self.objectKey(key)
	if err != nil {
		return err
	}
	buf := make([]byte, self.partSize)
	part, err := readPart(reader, buf)
	if err != nil {
		return err
	}
	if len(part) < self.partSize {
		_, err := self.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(self.bucket),
			Key:           objectKey,
			Body:          bytes.NewReader(part),
			ContentLength: aws.Int64(int64(len(part))),
			ContentType:   contentTypeOrNil(contentType),
		})
		return err
	}
	return self.putMultipart(ctx, objectKey, contentType, reader, buf)
}

func (self *S3Store) putMultipart(ctx context.Context, objectKey *string, contentType string, reader io.Reader, buf []byte) error {
	created, err := self.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(self.bucket),
		Key:         objectKey,
		ContentType: contentTypeOrNil(contentType),
	})
	if err != nil {
		return err
	}
	if err := self.uploadParts(ctx, objectKey, created.UploadId, reader, buf); err != nil {
		_, _ = self.client.AbortMultipartUpload(context.WithoutCancel(ctx), &s3.AbortMultipartUploadInput{
			Bucket: aws.String(self.bucket), Key: objectKey, UploadId: created.UploadId,
		})
		return err
	}
	return nil
}

func (self *S3Store) uploadParts(ctx context.Context, objectKey, uploadID *string, reader io.Reader, buf []byte) error {
	var completed []types.CompletedPart
	part := buf
	for number := int32(1); len(part) > 0; number++ {
		uploaded, err := self.client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:        aws.String(self.bucket),
			Key:           objectKey,
			UploadId:      uploadID,
			PartNumber:    aws.Int32(number),
			Body:          bytes.NewReader(part),
			ContentLength: aws.Int64(int64(len(part))),
		})
		if err != nil {
			return err
		}
		completed = append(completed, types.CompletedPart{ETag: uploaded.ETag, PartNumber: aws.Int32(number)})
		if part, err = readPart(reader, buf); err != nil {
			return err
		}
	}
	_, err := self.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(self.bucket),
		Key:             objectKey,
		UploadId:        uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: completed},
	})
	return err
}

func (self *S3Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey, err := self.objectKey(key)
	if err != nil {
		return nil, err
	}
	object, err := self.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(self.bucket), Key: objectKey})
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return object.Body, nil
}

func (self *S3Store) Exists(ctx context.Context, key string) (bool, error) {
	objectKey, err := self.objectKey(key)
	if err != nil {
		return false, err
	}
	_, err = self.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(self.bucket), Key: objectKey})
	if isNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func (self *S3Store) Delete(ctx context.Context, key string) error {
	objectKey, err := self.objectKey(key)
	if err != nil {
		return err
	}
	_, err = self.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(self.bucket), Key: objectKey})
	if isNotFound(err) {
		return nil
	}
	return err
}

func (self *S3Store) objectKey(key string) (*string, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	return aws.String(self.keyPrefix + uploadsPrefix + "/" + key), nil
}

func normalizePrefix(prefix string) string {
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		return ""
	}
	return prefix + "/"
}

func readPart(reader io.Reader, buf []byte) ([]byte, error) {
	n, err := io.ReadFull(reader, buf)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return buf[:n], nil
	}
	return buf[:n], err
}

func contentTypeOrNil(contentType string) *string {
	if contentType == "" {
		return nil
	}
	return aws.String(contentType)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var noSuchKey *types.NoSuchKey
	var notFound *types.NotFound
	if errors.As(err, &noSuchKey) || errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound")
}
