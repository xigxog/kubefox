package s3

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/xigxog/kubefox/kit"
)

type Client struct {
	ktx     kit.Kontext
	wrapped *minio.Client
}

func New(ktx kit.Kontext, dependency kit.Dependency) *Client {
	c, _ := minio.New("minio.default", &minio.Options{
		Transport: ktx.Transport(dependency),
	})

	return &Client{
		ktx:     ktx,
		wrapped: c,
	}
}

func (c *Client) MakeBucket(bucketName string, opts minio.MakeBucketOptions) error {
	return c.MakeBucketCtx(c.ktx.Context(), bucketName, opts)
}

func (c *Client) MakeBucketCtx(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return c.wrapped.MakeBucket(ctx, bucketName, opts)
}

func (c *Client) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return c.ListBucketsCtx(c.ktx.Context())
}

func (c *Client) ListBucketsCtx(ctx context.Context) ([]minio.BucketInfo, error) {
	return c.wrapped.ListBuckets(ctx)
}

func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return c.BucketExistsCtx(c.ktx.Context(), bucketName)
}

func (c *Client) BucketExistsCtx(ctx context.Context, bucketName string) (bool, error) {
	return c.wrapped.BucketExists(ctx, bucketName)
}

func (c *Client) RemoveBucket(bucketName string) error {
	return c.RemoveBucketCtx(c.ktx.Context(), bucketName)
}

func (c *Client) RemoveBucketCtx(ctx context.Context, bucketName string) error {
	return c.wrapped.RemoveBucket(ctx, bucketName)
}

func (c *Client) ListObjects(bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.ListObjectsCtx(c.ktx.Context(), bucketName, opts)
}

func (c *Client) ListObjectsCtx(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.wrapped.ListObjects(ctx, bucketName, opts)
}

func (c *Client) ListIncompleteUploads(bucketName, objectPrefix string, recursive bool) <-chan minio.ObjectMultipartInfo {
	return c.ListIncompleteUploadsCtx(c.ktx.Context(), bucketName, objectPrefix, recursive)
}

func (c *Client) ListIncompleteUploadsCtx(ctx context.Context, bucketName, objectPrefix string, recursive bool) <-chan minio.ObjectMultipartInfo {
	return c.wrapped.ListIncompleteUploads(ctx, bucketName, objectPrefix, recursive)
}
