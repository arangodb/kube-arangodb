package s3

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

const (
	// MaxChunkBytes contains the maximum number of bytes in a chunk (for Read/Write streaming operations)
	MaxChunkBytes = 1024 * 1024
)

// InputWriter abstracts the implementation of input writers.
type InputWriter interface {
	// WriteInput writes the next chunk into the bucket object.
	// Returns: moreDataAllowed, error
	WriteInput(ctx context.Context, chunk []byte, hasMore bool) (bool, error)
	// Close the underlying object
	Close() error
}

// OutputReader abstracts the implementation of output readers.
type OutputReader interface {
	// ReadOutput fetches the next chunk available for the bucket object with given info.
	// Returns: chunk, moreData, error
	// If an empty chunk is returned and moreData is true, more data is expected.
	ReadOutput(ctx context.Context) ([]byte, bool, error)
}

// NewBucketOutputReader create a BucketOutputReader
func NewBucketOutputReader(log logging.Logger, readCloser io.ReadCloser, encryptionKey []byte) (OutputReader, error) {
	var reader io.Reader
	if len(encryptionKey) > 0 {
		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			log.Err(err).Debug("aes.NewCipher failed")
			return nil, err
		}
		// If the key is unique for each ciphertext (aka object in the bucket), then it's ok to use a zero IV.
		var iv [aes.BlockSize]byte
		stream := cipher.NewOFB(block, iv[:])
		// Wrap the provided reader in a cipher reader
		reader = &cipher.StreamReader{S: stream, R: readCloser}
	} else {
		reader = readCloser
	}

	return &bucketOutputReader{
		Log:    log,
		Reader: reader,
		Closer: readCloser,
	}, nil
}

// bucketOutputReader implements OutputReader with a io.Reader and io.Closer
type bucketOutputReader struct {
	Log    logging.Logger
	Reader io.Reader
	Closer io.Closer
}

// ReadOutput fetches the next chunk available for the bucket object with given info.
// Returns: chunk, moreData, error
// If an empty chunk is returned and moreData is true, more data is expected.
func (r *bucketOutputReader) ReadOutput(ctx context.Context) ([]byte, bool, error) {
	buf := make([]byte, MaxChunkBytes)
	n, err := r.Reader.Read(buf)
	if err != nil {
		if err == io.EOF {
			// The client must close the reader
			if err := r.Closer.Close(); err != nil {
				r.Log.Err(err).Debug("reader.Close failed")
				// Continue
			}
			// Returns we are done
			return buf[:n], false, nil
		}
		r.Log.Err(err).Debug("reader.Read failed")
		return nil, false, err
	}
	// Return the buffer, indicating more data is expected
	return buf[:n], true, nil
}

// NewBucketInputWriter creates a BucketInputWriter
func NewBucketInputWriter(log logging.Logger, pipeWriter *io.PipeWriter, encryptionKey []byte, uploadDone *sync.Mutex) (InputWriter, error) {
	var writer io.Writer
	if len(encryptionKey) > 0 {
		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			log.Err(err).Debug("aes.NewCipher failed")
			return nil, err
		}
		// If the key is unique for each ciphertext (aka object in the bucket), then it's ok to use a zero IV.
		var iv [aes.BlockSize]byte
		stream := cipher.NewOFB(block, iv[:])
		// Wrap the provided writer in a cipher writer
		writer = &cipher.StreamWriter{S: stream, W: pipeWriter}
	} else {
		writer = pipeWriter
	}

	return &bucketInputWriter{
		Log:        log,
		Writer:     writer,
		Closer:     pipeWriter,
		UploadDone: uploadDone,
	}, nil
}

// bucketInputWriter implements InputWriter with a io.Writer and io.Closer
type bucketInputWriter struct {
	Log        logging.Logger
	Writer     io.Writer
	Closer     io.Closer
	UploadDone *sync.Mutex
}

// WriteInput writes the next chunk into the bucket object.
// Returns: moreDataAllowed, error
func (w *bucketInputWriter) WriteInput(ctx context.Context, chunk []byte, hasMore bool) (bool, error) {
	if len(chunk) > 0 {
		n, err := w.Writer.Write(chunk)
		if err != nil {
			w.Log.Err(err).Debug("writer.Write failed")
			return false, err
		}
		if n != len(chunk) {
			w.Log.Int("expected", len(chunk)).Int("written", n).Warn("Inconsistent amount of data written")
			return false, fmt.Errorf("Inconsistent amount of data written")
		}
	}
	if !hasMore {
		// Close writer, so the async code will stop
		if err := w.Close(); err != nil {
			w.Log.Err(err).Debug("Close failed")
			return false, err
		}
	}
	// Return
	return hasMore, nil
}

// Close the underlaying object and wait it's fully done
func (w *bucketInputWriter) Close() error {
	// Close writer, so the async code will stop
	if err := w.Closer.Close(); err != nil {
		w.Log.Err(err).Debug("writer.Close failed")
		return err
	}
	// Wait until the upload if fully done
	w.UploadDone.Lock()
	return nil
}
