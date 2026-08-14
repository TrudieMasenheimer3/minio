package cmd

import (
	"context"
)

// CompleteMultipartUpload - completes a multipart upload and invalidates the cache for the object.
func (c *cacheObjects) CompleteMultipartUpload(ctx context.Context, bucket, object, uploadID string, uploadedParts []CompletePart, opts ObjectOptions) (objInfo ObjectInfo, err error) {
	objInfo, err = c.ObjectLayer.CompleteMultipartUpload(ctx, bucket, object, uploadID, uploadedParts, opts)
	if err != nil {
		return objInfo, err
	}
	c.deleteCacheObject(ctx, bucket, object)
	return objInfo, nil
}