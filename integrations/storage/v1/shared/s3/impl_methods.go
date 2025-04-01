//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package s3

import (
	"context"
	"fmt"
	"io"
	goStrings "strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbCommon "github.com/arangodb-managed/apis/common/v1"
	pbStorage "github.com/arangodb-managed/integration-apis/bucket-service/v1"
	ipbCommon "github.com/arangodb-managed/integration-apis/common/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (s *s3impl) GetAPIVersion(context.Context, *pbCommon.Empty) (*pbCommon.Version, error) {
	return &pbCommon.Version{
		Major: pbStorage.APIMajorVersion,
		Minor: pbStorage.APIMinorVersion,
		Patch: pbStorage.APIPatchVersion,
	}, nil
}

// BucketExists checks if the specified bucket exists
func (s *s3impl) BucketExists(ctx context.Context, req *pbStorage.BucketRequest) (*pbCommon.YesOrNo, error) {
	log := logger.Str("func", "BucketExists")

	if len(req.Tags) > 0 {
		return nil, pbCommon.InvalidArgument("tags not supported")
	}

	// Check if the bucket exists
	if _, err := s.client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: util.NewType(s.bucket),
	}); err != nil {
		// See https://github.com/aws/aws-sdk-go/issues/2593
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket, "NotFound":
				return &pbCommon.YesOrNo{Result: false}, nil
			}
		}
		log.Err(err).Error("BucketExists - HeadBucketWithContext failed")
		return &pbCommon.YesOrNo{Result: false}, err
	}
	return &pbCommon.YesOrNo{Result: true}, nil
}

// CreateBucket creates a bucket
func (s *s3impl) CreateBucket(ctx context.Context, req *pbStorage.BucketRequest) (*pbCommon.Empty, error) {
	log := logger.Str("func", "CreateBucket")

	// Create bucket
	var createBucketConfiguration *s3.CreateBucketConfiguration
	if s.region != "" {
		// From AWS documentation:
		// Specifies the Region where the bucket will be created. If you don't specify
		// a Region, the bucket is created in the US East (N. Virginia) Region (us-east-1).
		createBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: util.NewType(s.region),
		}
	}
	const bucketACL = "private"
	if _, err := s.client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket:                    util.NewType(s.bucket),
		ACL:                       util.NewType(bucketACL),
		CreateBucketConfiguration: createBucketConfiguration,
	}); err != nil {
		log.Err(err).Error("Bucket create failed: CreateBucket")
		return nil, err
	}
	// Wait until bucket is created before finishing
	if err := s.client.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: util.NewType(s.bucket),
	}); err != nil {
		log.Err(err).Error("Bucket create failed: WaitUntilBucketExists")
		return nil, err
	}
	// Fill TagSet
	tagSet := make([]*s3.Tag, 0)
	for _, kv := range req.GetTags() {
		tagSet = append(tagSet, &s3.Tag{Key: util.NewType(kv.GetKey()), Value: util.NewType(kv.GetValue())})
	}
	if _, err := s.client.PutBucketTaggingWithContext(ctx, &s3.PutBucketTaggingInput{
		Bucket: util.NewType(s.bucket),
		Tagging: &s3.Tagging{
			TagSet: tagSet,
		},
	}); err != nil {
		log.Err(err).Error("Bucket create failed: PutBucketTaggingWithContext")
		return nil, err
	}
	log.Debug("Bucket created")
	return &pbCommon.Empty{}, nil

}

// DeleteBucket deletes a bucket
// Notice that this deletes all data contained in the bucket as well
func (s *s3impl) DeleteBucket(ctx context.Context, req *pbStorage.BucketRequest) (*pbCommon.Empty, error) {
	log := logger.Str("func", "DeleteBucket")

	if len(req.Tags) > 0 {
		return nil, pbCommon.InvalidArgument("tags not supported")
	}

	// DeleteBucket on AWS is allowed when there are no more files in the bucket, so delete them first (if any)
	if _, err := s.DeletePath(ctx, &pbStorage.PathRequest{Path: "."}); err != nil {
		log.Err(err).Error("Bucket delete failed: DeletePath")
		return nil, err
	}
	// Delete bucket
	if _, err := s.client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: util.NewType(s.bucket),
	}); err != nil {
		log.Err(err).Error("Bucket delete failed")
		return nil, err
	}
	// Wait until bucket is deleted before finishing
	if err := s.client.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: util.NewType(s.bucket),
	}); err != nil {
		log.Err(err).Error("Bucket delete failed: WaitUntilBucketNotExists")
		return nil, err
	}
	log.Debug("Bucket deleted")

	return &pbCommon.Empty{}, nil
}

// GetRepositoryURL get the URL needed to store/delete objects in a bucket
func (s *s3impl) GetRepositoryURL(_ context.Context, req *pbStorage.PathRequest) (*pbStorage.RepositoryURL, error) {
	// Check request fields
	path := req.GetPath()
	if path == "" {
		return nil, pbCommon.InvalidArgument("path missing")
	}
	return s.getRepositoryURL(path), nil
}

// DeletePath deletes the specified path (recursively) from the provided bucket
func (s *s3impl) DeletePath(ctx context.Context, req *pbStorage.PathRequest) (*pbCommon.Empty, error) {
	log := logger.Str("func", "DeletePath").Str("path", req.GetPath())

	path := req.GetPath()
	if path == "" {
		return nil, pbCommon.InvalidArgument("path cannot be empty")
	}

	counter := 0
	var continuationToken *string
	for {
		// Select object to delete
		objs, err := s.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket:            util.NewType(s.bucket),
			Prefix:            s.getPrefix(path),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			log.Err(err).Error("Bucket delete-path failed: ListObjectsV2WithContext")
			return nil, err
		}
		// Delete if we have selected any objects
		if len(objs.Contents) > 0 {
			counter += len(objs.Contents)
			// Create ObjectIdentifier collection
			objsIds := make([]*s3.ObjectIdentifier, 0, len(objs.Contents))
			for _, obj := range objs.Contents {
				objsIds = append(objsIds, &s3.ObjectIdentifier{
					Key: obj.Key,
				})
			}
			// Delete Objects
			if _, err := s.client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
				Bucket: objs.Name,
				Delete: &s3.Delete{
					Objects: objsIds,
				},
			}); err != nil {
				log.Err(err).Error("Bucket delete path failed")
			}
		}
		if !util.TypeOrDefault(objs.IsTruncated) {
			// We are done
			break
		}
		continuationToken = objs.NextContinuationToken
	}
	if counter > 0 {
		log.Int("counter", counter).Debug("DeletePath deleted objects")
	}
	return &pbCommon.Empty{}, nil
}

// GetPathSize provides the size in bytes for the specified path from the provided bucket
func (s *s3impl) GetPathSize(ctx context.Context, req *pbStorage.PathRequest) (*pbStorage.PathSize, error) {
	log := logger.Str("func", "DeletePath").Str("path", req.GetPath())

	// Check request fields
	path := req.GetPath()
	if path == "" {
		return nil, pbCommon.InvalidArgument("path missing")
	}

	// Get bucket info
	pSize, err := s.PathInfo(ctx, path)
	if err != nil {
		log.Err(err).Debug("PathInfo failed")
		return nil, err
	}

	return &pSize, nil
}

// PathInfo provides the bucket size in bytes and number of files in the bucket for the specified path from the provided bucket
// Specify path as "." to indicate the root folder
func (s *s3impl) PathInfo(ctx context.Context, path string) (pbStorage.PathSize, error) {
	log := logger.Str("func", "PathInfo").Str("path", path)
	if path == "" {
		return pbStorage.PathSize{}, pbCommon.InvalidArgument("path cannot be empty")
	}
	var pSize int64
	var numFiles int32
	var continuationToken *string
	prefix := s.getPrefix(path)
	folderCounter := NewBucketFolderCounter(*prefix)
	for {
		objs, err := s.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket:            util.NewType(s.bucket),
			Prefix:            prefix,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			log.Err(err).Error("Bucket path size failed: ListObjectsV2WithContext")
			return pbStorage.PathSize{}, err
		}
		for _, obj := range objs.Contents {
			folderCounter.AddObject(util.TypeOrDefault(obj.Key))
			numFiles++
			size := util.TypeOrDefault(obj.Size)
			pSize += size
		}
		if !util.TypeOrDefault(objs.IsTruncated) {
			// We are done
			break
		}
		continuationToken = objs.NextContinuationToken
	}

	return pbStorage.PathSize{
		SizeInBytes:     uint64(pSize),
		NumberOfFiles:   uint32(numFiles),
		NumberOfFolders: uint32(folderCounter.GetFolderCount()),
	}, nil
}

// createReader creates a reader to return the content of a bucket object.
func (s *s3impl) createReader(ctx context.Context, path string) (OutputReader, error) {
	log := logger.Str("func", "CreateReader").Str("path", path)
	if path == "" {
		return nil, pbCommon.InvalidArgument("path cannot be empty")
	}

	obj, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: util.NewType(s.bucket),
		Key:    s.getPrefix(path),
	})
	if err != nil {
		log.Err(err).Debug("GetObjectWithContext failed")
		return nil, err
	}

	readerLog := logger.Str("path", path).Str("ctx", "output-reader")

	return NewBucketOutputReader(readerLog, obj.Body, nil)
}

// ReadObject opens an object in the bucket and streams the existing data from the object into the client
func (s *s3impl) ReadObject(req *pbStorage.PathRequest, server pbStorage.BucketService_ReadObjectServer) error {
	log := logger.Str("func", "ReadObject").Str("path", req.GetPath())
	ctx := server.Context()
	path := req.GetPath()
	if path == "" {
		return pbCommon.InvalidArgument("path missing")
	}

	funcResult := util.NewType("fail")
	defer func() {
		// Log when we are terminated this call
		log.Str("result", *funcResult).Debug("ReadObject terminated (defer)")
	}()
	// Create the reader
	reader, err := s.createReader(ctx, path)
	if err != nil {
		log.Err(err).Debug("CreateReader failed")
		return err
	}

	for {
		// Read next chunk
		chunk, moreData, err := reader.ReadOutput(ctx)
		if err != nil {
			log.Err(err).Debug("ReadOutput failed")
			return err
		}
		// Send chunk to caller
		if err := server.Send(&pbStorage.ReadObjectChunk{
			Chunk: chunk,
		}); err != nil {
			log.Err(err).Debug("Failed to send ReadObjectChunk")
			return err
		}
		// Stop if needed
		if !moreData {
			// No more data expected
			log.Debug("ReadOutput done")
			funcResult = util.NewType("ok")
			return nil
		}
	}
}

// createWriter creates a writer to store chunks in the provided bucket object.
func (s *s3impl) createWriter(ctx context.Context, path string, tags ipbCommon.KeyValuePairList) (InputWriter, error) {
	log := logger.Str("func", "createWriter").Str("path", path)
	if path == "" {
		return nil, pbCommon.InvalidArgument("path cannot be empty")
	}
	info, err := s.getObjectInfo(ctx, path)
	if err != nil {
		log.Err(err).Debug("getObjectInfo failed")
		return nil, err
	}
	// Do not write to locked objects
	if info.IsLocked {
		log.Debug("Object is locked")
		return nil, pbCommon.PreconditionFailed("Object is locked")
	}
	prefix := s.getPrefix(path)
	objPath := fmt.Sprintf("%s/%s", s.bucket, *prefix)
	queryTemp := fmt.Sprintf("%s.%s.tmp", objPath, newID())
	g, lctx := errgroup.WithContext(ctx)
	r, w := io.Pipe()
	// Check if blob already exists, if so make sure the original content is preserved
	if !info.Exists {
		// Copy tags into info object
		info.Tags = tags
	} else {
		// Copy original file to temp file
		if _, err := s.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
			Bucket:     util.NewType(s.bucket),
			Key:        util.NewType(queryTemp),
			CopySource: util.NewType(objPath),
		}); err != nil {
			log.Err(err).Debug("CopyObjectWithContext failed")
			return nil, err
		}
		// Append temp data (async to the write, so we do not need to buffer the complete document) back to original
		g.Go(func() error {
			obj, err := s.client.GetObjectWithContext(lctx, &s3.GetObjectInput{
				Bucket: util.NewType(s.bucket),
				Key:    util.NewType(queryTemp),
			})
			if err != nil {
				log.Err(err).Debug("GetObjectWithContext failed")
				w.CloseWithError(err)
				return err
			}
			// Write the existing data
			if _, err := io.Copy(w, obj.Body); err != nil {
				log.Err(err).Debug("io.Copy failed")
				w.CloseWithError(err)
				return err
			}
			if err := obj.Body.Close(); err != nil {
				log.Err(err).Debug("Body.Close failed")
				w.CloseWithError(err)
				return err
			}
			return nil
		})
	}
	// Create uploader
	uploader := s3manager.NewUploaderWithClient(s.client)
	uploadDone := sync.Mutex{}
	uploadDone.Lock()
	// Upload async
	go func() {
		// Unlock the upload done to indicate we are fully done
		defer uploadDone.Unlock()
		if _, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket: util.NewType(s.bucket),
			Key:    prefix,
			Body:   r,
		}); err != nil {
			log.Err(err).Debug("UploadWithContext failed")
			r.CloseWithError(err)
			return
		}
		// Set tags (if needed)
		if len(info.Tags) > 0 {
			// Fill TagSet
			tagSet := make([]*s3.Tag, 0, len(info.Tags))
			for _, kv := range info.Tags {
				tagSet = append(tagSet, &s3.Tag{Key: util.NewType(kv.GetKey()), Value: util.NewType(kv.GetValue())})
			}
			if _, err := s.client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
				Bucket: util.NewType(s.bucket),
				Key:    prefix,
				Tagging: &s3.Tagging{
					TagSet: tagSet,
				},
			}); err != nil {
				log.Err(err).Warn("PutObjectTaggingWithContext failed (tagging)")
				r.CloseWithError(err)
				return
			}
		}
		// Lock object
		if err := s.lockObject(ctx, log, path); err != nil {
			log.Err(err).Warn("lockObject failed")
			r.CloseWithError(err)
			return
		}
	}()
	// Wait until the original file has been written with the temp data (if needed)
	if err := g.Wait(); err != nil {
		log.Err(err).Debug("Group wait failed")
		return nil, err
	}
	// Delete temp file (if needed)
	if info.Exists {
		if _, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: util.NewType(s.bucket),
			Key:    util.NewType(queryTemp),
		}); err != nil {
			log.Err(err).Debug("DeleteObjectWithContext failed")
			return nil, err
		}
	}
	// Return writer, which should be closed by caller
	return NewBucketInputWriter(log, w, nil, &uploadDone)
}

// WriteObject creates or opens an object in the bucket and allows the client to stream (additional) data into the object
func (s *s3impl) WriteObject(server pbStorage.BucketService_WriteObjectServer) error {
	ctx := server.Context()
	log := logger.Str("func", "WriteObject")

	var (
		pathReq *pbStorage.PathRequest
		path    string
		writer  InputWriter
	)
	funcResult := util.NewType("fail")
	defer func() {
		log.Str("result", *funcResult).Debug("WriteObject terminated (defer)")
	}()

	closeLoop := false
	for {
		// Send the control message
		if err := server.Send(&pbStorage.WriteObjectControl{
			AllowMoreOutput: !closeLoop,
			MaxChunkBytes:   MaxChunkBytes,
		}); err != nil {
			log.Err(err).Debug("Failed to send WriteObjectControl message")
			return err
		}

		// Wait for output message
		msg, err := server.Recv()
		if closeLoop || err == io.EOF || pbCommon.IsCanceled(err) {
			// Connection terminated
			funcResult = util.NewType("ok")
			// Unregister writer (if needed)
			if writer != nil {
				found := false
				s.registeredWritersMutex.Lock()
				for i, w := range s.registeredWriters {
					if w == &writer {
						// Remove writer
						found = true
						s.registeredWriters = append(s.registeredWriters[:i], s.registeredWriters[i+1:]...)
						// break the loop
						break
					}
				}
				s.registeredWritersMutex.Unlock()
				if !found {
					log.Debug("Writer not found in registeredWriters, so cannot remove")
				}
			}
			return nil
		} else if err != nil {
			log.Err(err).Debug("Failed to read WriteObjectControl message")
			return err
		}
		if pathReq == nil {
			// First received message, check path
			pathReq = msg.GetPath()
			path = pathReq.GetPath()
			log = log.Str("path", path)
			if path == "" {
				log.Debug("path missing")
				return pbCommon.InvalidArgument("path missing")
			}
			// Create writer
			writer, err = s.createWriter(ctx, path, nil)
			if err != nil {
				if !pbCommon.IsUnavailable(err) {
					log.Err(err).Debug("createWriter failed")
					return err
				}
				// If Unavailable for this provider, the normal Write will be used, otherwise the writer will be used
			}
			// Register writer (if needed)
			if writer != nil {
				s.registeredWritersMutex.Lock()
				s.registeredWriters = append(s.registeredWriters, &writer)
				s.registeredWritersMutex.Unlock()
			}
		}
		// Dependent if the writer is available we will use that, or the normal write.
		var moreDataAllowed bool
		if writer == nil {
			// Pass chunk to provider
			if moreDataAllowed, err = s.write(ctx, path, msg.GetChunk(), msg.GetHasMore(), nil); err != nil {
				log.Debug("bucketProvider.Write failed")
				return err
			}
		} else {
			// Use writer
			if moreDataAllowed, err = writer.WriteInput(ctx, msg.GetChunk(), msg.GetHasMore()); err != nil {
				log.Debug("writer.WriteInput failed")
				return err
			}
		}
		// Should we continue?
		if !msg.GetHasMore() || !moreDataAllowed {
			// We're done (or no more data allowed)
			// Make sure to send last message
			closeLoop = true
		}
	}
}

// write stores the given chunk in the provided bucket object.
// Returns: moreDataAllowed, error
func (s *s3impl) write(ctx context.Context, path string, chunk []byte, hasMore bool, tags ipbCommon.KeyValuePairList) (bool, error) {
	log := logger.Str("func", "Write").Str("path", path)
	if path == "" {
		return false, pbCommon.InvalidArgument("path cannot be empty")
	}
	info, err := s.getObjectInfo(ctx, path)
	if err != nil {
		log.Err(err).Debug("getObjectInfo failed")
		return false, err
	}
	// Do not write to locked objects
	if info.IsLocked {
		log.Debug("Object is locked")
		return false, pbCommon.PreconditionFailed("Object is locked")
	}
	prefix := s.getPrefix(path)
	//objPath := fmt.Sprintf("%s/%s", s.bucketName, *prefix)
	tempPath := fmt.Sprintf("%s.%s.tmp", *prefix, newID())
	tempPathFull := fmt.Sprintf("%s/%s", s.bucket, tempPath)

	// Create blob (if needed)
	if !info.Exists {
		// Create
		if _, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket: util.NewType(s.bucket),
			Key:    prefix,
			Body:   aws.ReadSeekCloser(goStrings.NewReader(string(chunk))),
		}); err != nil {
			log.Err(err).Debug("PutObjectWithContext failed")
			return false, err
		}
		// Add tags (if needed)
		if len(tags) > 0 {
			// Fill TagSet
			tagSet := make([]*s3.Tag, 0, len(tags))
			for _, kv := range tags {
				tagSet = append(tagSet, &s3.Tag{Key: util.NewType(kv.GetKey()), Value: util.NewType(kv.GetValue())})
			}
			if _, err := s.client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
				Bucket: util.NewType(s.bucket),
				Key:    prefix,
				Tagging: &s3.Tagging{
					TagSet: tagSet,
				},
			}); err != nil {
				log.Err(err).Debug("PutObjectTaggingWithContext failed (tagging)")
				return false, err
			}
		}
		// Lock if no more data is expected
		if !hasMore {
			if err := s.lockObject(ctx, log, path); err != nil {
				log.Err(err).Debug("lockObject failed")
				return false, err
			}
		}
		return hasMore, nil
	}
	// Append data (async to the write, so we do not need to buffer the complete document)
	r, w := io.Pipe()
	go func() {
		obj, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: util.NewType(s.bucket),
			Key:    prefix,
		})
		if err != nil {
			w.CloseWithError(err)
			return
		}
		// Write the existing data
		if _, err := io.Copy(w, obj.Body); err != nil {
			w.CloseWithError(err)
			return
		}
		obj.Body.Close()
		// Add the new data
		if _, err := io.Copy(w, goStrings.NewReader(string(chunk))); err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}()
	// Create uploader
	uploader := s3manager.NewUploaderWithClient(s.client)
	// Write combined (in temp file, not to original, because AWS will first wipe the original and then attempt to write the new one)
	if _, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: util.NewType(s.bucket),
		Key:    util.NewType(tempPath),
		Body:   r,
	}); err != nil {
		r.CloseWithError(err)
		log.Err(err).Debug("UploadWithContext failed")
		return false, err
	}
	// Move temp to destination
	if _, err := s.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     util.NewType(s.bucket),
		Key:        prefix,
		CopySource: util.NewType(tempPathFull),
	}); err != nil {
		log.Err(err).Debug("CopyObjectWithContext failed")
		return false, err
	}
	// Restore tags (if needed)
	if len(info.Tags) > 0 {
		// Fill TagSet
		tagSet := make([]*s3.Tag, 0, len(info.Tags))
		for _, kv := range info.Tags {
			tagSet = append(tagSet, &s3.Tag{Key: util.NewType(kv.GetKey()), Value: util.NewType(kv.GetValue())})
		}
		if _, err := s.client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
			Bucket: util.NewType(s.bucket),
			Key:    prefix,
			Tagging: &s3.Tagging{
				TagSet: tagSet,
			},
		}); err != nil {
			log.Err(err).Debug("PutObjectTaggingWithContext failed (tagging restore)")
			return false, err
		}
	}
	// Delete temp file
	if _, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: util.NewType(s.bucket),
		Key:    util.NewType(tempPath),
	}); err != nil {
		log.Err(err).Debug("DeleteObjectWithContext failed")
		return false, err
	}
	// Lock if no more data is expected
	if !hasMore {
		if err := s.lockObject(ctx, log, path); err != nil {
			log.Err(err).Debug("lockObject failed")
			return false, err
		}
	}
	return hasMore, nil
}

// GetObjectInfo provides information for the specified object from the provided bucket
// A Not-Found error is returned if the object cannot be found
func (s *s3impl) GetObjectInfo(ctx context.Context, req *pbStorage.PathRequest) (*pbStorage.ObjectInfo, error) {
	log := logger.Str("func", "GetObjectInfo").Str("path", req.GetPath())

	// Check request fields
	path := req.GetPath()
	if path == "" {
		return nil, pbCommon.InvalidArgument("path missing")
	}
	result, err := s.getObjectInfo(ctx, path)
	if err != nil {
		log.Err(err).Debug("getObjectInfo failed")
		return nil, err
	}
	if !result.Exists {
		return nil, pbCommon.NotFound(path)
	}

	lastUpdatedAt := timestamppb.New(result.LastUpdatedAt)

	return &pbStorage.ObjectInfo{
		IsLocked:      result.IsLocked,
		SizeInBytes:   result.SizeInBytes,
		LastUpdatedAt: lastUpdatedAt,
	}, nil
}

func (s *s3impl) getObjectInfo(ctx context.Context, path string) (ObjectInfo, error) {
	log := logger.Str("func", "getObjectInfo").Str("path", path)
	// Get properties
	prefix := s.getPrefix(path)
	obj, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: util.NewType(s.bucket),
		Key:    prefix,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey, "NotFound":
				return ObjectInfo{Exists: false}, nil
			}
		}
		log.Err(err).Debug("blobURL.GetProperties failed")
		return ObjectInfo{}, err
	}
	tagging, err := s.client.GetObjectTaggingWithContext(ctx, &s3.GetObjectTaggingInput{
		Bucket: util.NewType(s.bucket),
		Key:    prefix,
	})
	if err != nil {
		log.Err(err).Debug("GetObjectTaggingWithContext failed")
		return ObjectInfo{}, err
	}
	// Get the locked value
	isLocked := false
	ts := tagging.TagSet
	for _, t := range ts {
		if util.TypeOrDefault(t.Key) == LockedKey {
			isLocked = util.TypeOrDefault(t.Value) == LockedValue
			break
		}
	}
	// Convert the tags
	tags := make(ipbCommon.KeyValuePairList, 0, len(ts))
	for _, t := range ts {
		if util.TypeOrDefault(t.Key) == LockedKey {
			// Filter out the LockedKey, this is for internal use only
			continue
		}
		tags.UpsertPair(util.TypeOrDefault(t.Key), util.TypeOrDefault(t.Value))
	}
	// Return result
	return ObjectInfo{
		Exists:        true,
		IsLocked:      isLocked,
		SizeInBytes:   uint64(util.TypeOrDefault(obj.ContentLength)),
		LastUpdatedAt: util.TypeOrDefault(obj.LastModified),
		Tags:          tags,
	}, nil
}

// Se the lock for the provided bucket object
func (s *s3impl) lockObject(ctx context.Context, log logging.Logger, path string) error {
	info, err := s.getObjectInfo(ctx, path)
	if err != nil {
		log.Err(err).Debug("getObjectInfo failed")
		return err
	}
	if info.IsLocked {
		log.Debug("Object is locked")
		return pbCommon.AlreadyExists("Lock already exists")
	}
	tagSet := []*s3.Tag{
		{Key: util.NewType(LockedKey), Value: util.NewType(LockedValue)},
	}
	// Merge existing metadata
	for _, kv := range info.Tags {
		tagSet = append(tagSet, &s3.Tag{Key: util.NewType(kv.GetKey()), Value: util.NewType(kv.GetValue())})
	}
	if _, err := s.client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
		Bucket: util.NewType(s.bucket),
		Key:    s.getPrefix(path),
		Tagging: &s3.Tagging{
			TagSet: tagSet,
		},
	}); err != nil {
		log.Err(err).Debug("PutObjectTaggingWithContext failed")
		return err
	}
	return nil
}
