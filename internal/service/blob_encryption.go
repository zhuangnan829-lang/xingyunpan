package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/storage"
)

const blobEncryptionMagic = "XYPENC1:"

const blobEncryptionHeaderSize = len(blobEncryptionMagic) + aes.BlockSize

type StoragePolicyEncryptionConfig struct {
	Enabled   bool
	KeyID     string
	MasterKey []byte
}

func (r storagePolicyRuntime) EncryptionConfig() (*StoragePolicyEncryptionConfig, error) {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil || !policy.EnableEncryption {
		return &StoragePolicyEncryptionConfig{}, err
	}

	keyID := strings.TrimSpace(policy.EncryptionKeyID)
	if keyID == "" {
		keyID = "default-master-key"
	}
	return &StoragePolicyEncryptionConfig{Enabled: true, KeyID: keyID}, nil
}

func readPhysicalBlob(stor storage.Storage, physical *model.PhysicalFile, resolver MasterKeyResolver) (io.ReadCloser, error) {
	if stor == nil {
		return nil, fmt.Errorf("storage is not available")
	}
	if physical == nil {
		return nil, fmt.Errorf("physical file is missing")
	}

	reader, err := stor.Read(physical.StoragePath)
	if err != nil {
		return nil, err
	}
	if !physical.Encrypted {
		return reader, nil
	}

	decrypted, err := decryptBlobReader(reader, physical.EncryptionKeyID, resolver)
	if err != nil {
		_ = reader.Close()
		return nil, err
	}
	return decrypted, nil
}

func savePhysicalBlob(stor storage.Storage, reader io.Reader, storagePath string, encryption *StoragePolicyEncryptionConfig, resolver MasterKeyResolver) error {
	if stor == nil {
		return fmt.Errorf("storage is not available")
	}
	if encryption != nil && encryption.Enabled {
		encrypted, err := encryptBlobReaderWithMasterKey(reader, encryption.KeyID, encryption.MasterKey, resolver)
		if err != nil {
			return err
		}
		reader = encrypted
	}
	return stor.Save(reader, storagePath)
}

func encryptionKeyID(config *StoragePolicyEncryptionConfig) string {
	if config == nil || !config.Enabled {
		return ""
	}
	return strings.TrimSpace(config.KeyID)
}

func prepareBlobEncryptionConfig(config *StoragePolicyEncryptionConfig, resolver MasterKeyResolver) (*StoragePolicyEncryptionConfig, error) {
	if config == nil || !config.Enabled {
		return config, nil
	}
	key, status, err := resolveBlobMasterKey(resolver)
	if err != nil {
		source := ""
		if status != nil {
			source = status.Source
		}
		if source != "" {
			return nil, fmt.Errorf("master key is not available (%s): %w", source, err)
		}
		return nil, fmt.Errorf("master key is not available: %w", err)
	}
	prepared := *config
	prepared.MasterKey = append([]byte{}, key...)
	return &prepared, nil
}

func physicalBlobStoredSize(fileSize int64, config *StoragePolicyEncryptionConfig) int64 {
	if fileSize <= 0 {
		return 0
	}
	if config != nil && config.Enabled {
		return fileSize + int64(blobEncryptionHeaderSize)
	}
	return fileSize
}

func encryptBlobReader(reader io.Reader, keyID string, resolver MasterKeyResolver) (io.Reader, error) {
	return encryptBlobReaderWithMasterKey(reader, keyID, nil, resolver)
}

func encryptBlobReaderWithMasterKey(reader io.Reader, keyID string, masterKey []byte, resolver MasterKeyResolver) (io.Reader, error) {
	nonce := make([]byte, aes.BlockSize)
	if _, err := rand.Read(nonce); err != nil {
		for i := range nonce {
			nonce[i] = byte(i + 1)
		}
	}

	key, err := deriveBlobEncryptionKeyWithMasterKey(keyID, masterKey, resolver)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	header := append([]byte(blobEncryptionMagic), nonce...)
	stream := cipher.NewCTR(block, nonce)
	encrypted := &cipher.StreamReader{S: stream, R: reader}
	return io.MultiReader(bytes.NewReader(header), encrypted), nil
}

func decryptBlobReader(reader io.ReadCloser, keyID string, resolver MasterKeyResolver) (io.ReadCloser, error) {
	header := make([]byte, len(blobEncryptionMagic)+aes.BlockSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, fmt.Errorf("read encrypted blob header failed: %w", err)
	}
	if string(header[:len(blobEncryptionMagic)]) != blobEncryptionMagic {
		return nil, fmt.Errorf("encrypted blob header is invalid")
	}

	key, err := deriveBlobEncryptionKey(keyID, resolver)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	nonce := header[len(blobEncryptionMagic):]
	stream := cipher.NewCTR(block, nonce)
	return &streamReadCloser{
		reader: &cipher.StreamReader{S: stream, R: reader},
		closer: reader,
	}, nil
}

func deriveBlobEncryptionKey(keyID string, resolver MasterKeyResolver) ([32]byte, error) {
	return deriveBlobEncryptionKeyWithMasterKey(keyID, nil, resolver)
}

func deriveBlobEncryptionKeyWithMasterKey(keyID string, masterKey []byte, resolver MasterKeyResolver) ([32]byte, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		keyID = "default-master-key"
	}
	if len(masterKey) == 0 {
		resolved, status, err := resolveBlobMasterKey(resolver)
		if err != nil {
			source := ""
			if status != nil {
				source = status.Source
			}
			if source != "" {
				return [32]byte{}, fmt.Errorf("master key is not available (%s): %w", source, err)
			}
			return [32]byte{}, fmt.Errorf("master key is not available: %w", err)
		}
		masterKey = resolved
	}
	material := bytes.Join([][]byte{masterKey, []byte(keyID)}, []byte{0})
	return sha256.Sum256(material), nil
}

func resolveBlobMasterKey(resolver MasterKeyResolver) ([]byte, *MasterKeyStatusPayload, error) {
	if resolver == nil {
		return nil, nil, fmt.Errorf("master key resolver is not available")
	}
	return resolver.ResolveMasterKey()
}

type streamReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (r *streamReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *streamReadCloser) Close() error {
	return r.closer.Close()
}
