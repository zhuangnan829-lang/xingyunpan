import CryptoJS from 'crypto-js';

/**
 * Calculate SHA256 hash of a file
 * Supports large files by processing in chunks to avoid memory overflow
 * @param file - The file to calculate hash for
 * @param onProgress - Optional callback for progress updates (0-100)
 * @returns Promise<string> - The SHA256 hash in hex format
 */
export async function calculateFileHash(
  file: File,
  onProgress?: (progress: number) => void
): Promise<string> {
  const chunkSize = 2 * 1024 * 1024; // 2MB chunks for hash calculation
  const chunks = Math.ceil(file.size / chunkSize);
  const hasher = CryptoJS.algo.SHA256.create();

  for (let i = 0; i < chunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    // Read chunk as ArrayBuffer
    const arrayBuffer = await readChunkAsArrayBuffer(chunk);

    // Convert ArrayBuffer to WordArray for crypto-js
    const wordArray = CryptoJS.lib.WordArray.create(arrayBuffer as any);

    // Update hasher with chunk
    hasher.update(wordArray);

    // Report progress if callback provided
    if (onProgress) {
      const progress = Math.floor(((i + 1) / chunks) * 100);
      onProgress(progress);
    }
  }

  // Finalize hash and return as hex string
  return hasher.finalize().toString();
}

/**
 * Read a blob/chunk as ArrayBuffer
 * @param blob - The blob to read
 * @returns Promise<ArrayBuffer>
 */
function readChunkAsArrayBuffer(blob: Blob): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = () => {
      if (reader.result instanceof ArrayBuffer) {
        resolve(reader.result);
      } else {
        reject(new Error('Failed to read chunk as ArrayBuffer'));
      }
    };

    reader.onerror = () => {
      reject(reader.error);
    };

    reader.readAsArrayBuffer(blob);
  });
}
