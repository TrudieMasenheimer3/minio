package cmd

import (
	"bytes"
	"context"
	"testing"
)

func TestMultipartOverwriteETagConsistency(t *testing.T) {
	ExecObjectLayerTest(t, testMultipartOverwriteETagConsistency)
}

func testMultipartOverwriteETagConsistency(obj ObjectLayer, instanceType string, t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket-etag"
	object := "test-object-etag"

	err := obj.MakeBucketWithLocation(ctx, bucket, BucketOptions{})
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	// 1. Put initial object
	_, err = obj.PutObject(ctx, bucket, object, mustGetPutObjReader(t, bytes.NewReader([]byte("old-content")), 11, "", ""), ObjectOptions{})
	if err != nil {
		t.Fatalf("Failed to put initial object: %v", err)
	}

	// Get initial ETag
	objInfoOld, err := obj.GetObjectInfo(ctx, bucket, object, ObjectOptions{})
	if err != nil {
		t.Fatalf("Failed to get object info: %v", err)
	}
	oldETag := objInfoOld.ETag

	// 2. Perform Multipart Upload to overwrite the same object
	res, err := obj.NewMultipartUpload(ctx, bucket, object, ObjectOptions{})
	if err != nil {
		t.Fatalf("Failed to initiate multipart upload: %v", err)
	}

	part1, err := obj.PutObjectPart(ctx, bucket, object, res.UploadID, 1, mustGetPutObjReader(t, bytes.NewReader([]byte("new-content-part-1")), 18, "", ""), ObjectOptions{})
	if err != nil {
		t.Fatalf("Failed to upload part 1: %v", err)
	}

	parts := []CompletePart{
		{PartNumber: 1, ETag: part1.ETag},
	}

	_, err = obj.CompleteMultipartUpload(ctx, bucket, object, res.UploadID, parts, ObjectOptions{})
	if err != nil {
		t.Fatalf("Failed to complete multipart upload: %v", err)
	}

	// 3. Call GetObjectInfo immediately in a loop to verify no stale ETag is returned
	for i := 0; i < 10; i++ {
		objInfoNew, err := obj.GetObjectInfo(ctx, bucket, object, ObjectOptions{})
		if err != nil {
			t.Fatalf("Failed to get object info on iteration %d: %v", i, err)
		}
		if objInfoNew.ETag == oldETag {
			t.Fatalf("Stale ETag returned on iteration %d: %s", i, oldETag)
		}
	}
}