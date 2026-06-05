package service

import (
	"fmt"
	"strings"

	"xingyunpan-v2/pkg/storage"
)

type preAllocatingStorage interface {
	PreAllocate(relativePath string, size int64) error
	Delete(relativePath string) error
}

func maybePreAllocateBlob(runtime storagePolicyRuntime, stor preAllocatingStorage, storagePath string, size int64) error {
	if stor == nil || strings.TrimSpace(storagePath) == "" || size <= 0 {
		return nil
	}
	enabled, err := runtime.PreAllocateEnabled()
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	if err := stor.PreAllocate(storagePath, size); err != nil {
		_ = stor.Delete(storagePath)
		return fmt.Errorf("pre-allocate storage space failed: %w", err)
	}
	return nil
}

func maybePreAllocateMultipartBlob(runtime storagePolicyRuntime, stor storage.MultipartStorage, storagePath string, size int64) error {
	preAlloc, ok := stor.(preAllocatingStorage)
	if !ok {
		return nil
	}
	return maybePreAllocateBlob(runtime, preAlloc, storagePath, size)
}
