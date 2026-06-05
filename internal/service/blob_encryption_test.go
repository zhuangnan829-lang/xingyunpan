package service

import (
	"io"
	"strings"
	"testing"
)

type memoryReadCloser struct {
	*strings.Reader
}

func (r memoryReadCloser) Close() error {
	return nil
}

type staticMasterKeyResolver struct {
	key []byte
}

func (r staticMasterKeyResolver) ResolveMasterKey() ([]byte, *MasterKeyStatusPayload, error) {
	return r.key, masterKeyStatus("test", "test", r.key), nil
}

func (r staticMasterKeyResolver) Status() *MasterKeyStatusPayload {
	return masterKeyStatus("test", "test", r.key)
}

func TestBlobEncryptionRoundTrip(t *testing.T) {
	plain := "hello encrypted blob"
	resolver := staticMasterKeyResolver{key: []byte("test-master-key")}
	encryptedReader, err := encryptBlobReader(strings.NewReader(plain), "key-a", resolver)
	if err != nil {
		t.Fatalf("encryptBlobReader failed: %v", err)
	}
	encrypted, err := io.ReadAll(encryptedReader)
	if err != nil {
		t.Fatalf("read encrypted data failed: %v", err)
	}
	if string(encrypted) == plain {
		t.Fatal("expected encrypted data to differ from plaintext")
	}
	if !strings.HasPrefix(string(encrypted), blobEncryptionMagic) {
		t.Fatal("expected encrypted blob header")
	}

	reader, err := decryptBlobReader(memoryReadCloser{Reader: strings.NewReader(string(encrypted))}, "key-a", resolver)
	if err != nil {
		t.Fatalf("decryptBlobReader failed: %v", err)
	}
	defer reader.Close()

	decrypted, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read decrypted data failed: %v", err)
	}
	if string(decrypted) != plain {
		t.Fatalf("decrypted content mismatch: %q", decrypted)
	}
}

func TestDecryptBlobRejectsInvalidHeader(t *testing.T) {
	if _, err := decryptBlobReader(memoryReadCloser{Reader: strings.NewReader("plain data")}, "key-a", staticMasterKeyResolver{key: []byte("test-master-key")}); err == nil {
		t.Fatal("expected invalid encrypted header to fail")
	}
}
