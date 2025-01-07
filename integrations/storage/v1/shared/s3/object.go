package s3

import (
	"time"

	pbCommon "github.com/arangodb-managed/integration-apis/common/v1"
)

// ObjectInfo is the response type for the GetObjectInfo method
type ObjectInfo struct {
	// Indicates if the object exists
	Exists bool
	// Indicates if the object is locked
	// This info is only relevant if the object exists
	IsLocked bool
	// Indicates the size of the object in bytes
	// This info is only relevant if the object exists
	SizeInBytes uint64
	// The timestamp this object has last been modified
	LastUpdatedAt time.Time
	// The tags associated with this object, if any
	Tags pbCommon.KeyValuePairList
}
