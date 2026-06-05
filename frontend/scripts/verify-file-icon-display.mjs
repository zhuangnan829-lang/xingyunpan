import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(import.meta.dirname, '..');

function read(relativePath) {
  return readFileSync(resolve(root, relativePath), 'utf8');
}

function assertContains(source, needle, label) {
  if (!source.includes(needle)) {
    throw new Error(`${label} missing expected source: ${needle}`);
  }
}

const fileItem = read('src/components/FileItem/index.vue');
assertContains(fileItem, "const displayIcon = computed(() => currentItem.value?.display_icon || '')", 'FileItem display_icon priority');
assertContains(fileItem, 'v-else-if="displayIcon"', 'FileItem custom icon branch');
assertContains(fileItem, '{{ displayIcon }}', 'FileItem custom icon render');
assertContains(fileItem, "const displayIconTint = computed(() => currentItem.value?.display_icon_tint || '#64748b')", 'FileItem display_icon_tint');
assertContains(fileItem, 'const displayIconLabel = computed(() => currentItem.value?.display_icon_label || displayName.value)', 'FileItem display_icon_label');

const fileList = read('src/components/FileList/index.vue');
assertContains(fileList, "const getDisplayIcon = (item: any) => item.display_icon || ''", 'FileList display_icon helper');
assertContains(fileList, "const getDisplayIconTint = (item: any) => item.display_icon_tint || '#64748b'", 'FileList display_icon_tint helper');
assertContains(fileList, 'v-else-if="getDisplayIcon(row)"', 'FileList table custom icon branch');
assertContains(fileList, '{{ getDisplayIcon(row) }}', 'FileList table custom icon render');
assertContains(fileList, '<FileItem', 'FileList grid delegates to FileItem');
assertContains(fileList, ':file="item.isFolder ? undefined : item"', 'FileList grid passes file display_icon payload');
assertContains(fileList, ':folder="item.isFolder ? item : undefined"', 'FileList grid passes folder display_icon payload');
assertContains(fileList, 'v-else-if="getDisplayIcon(item)"', 'FileList gallery custom icon branch');
assertContains(fileList, '{{ getDisplayIcon(item) }}', 'FileList gallery custom icon render');

const fileApi = read('src/api/file.ts');
assertContains(fileApi, 'display_icon: item.display_icon ||', 'File API maps display_icon');
assertContains(fileApi, 'display_icon_tint: item.display_icon_tint ||', 'File API maps display_icon_tint');
assertContains(fileApi, 'display_icon_label: item.display_icon_label ||', 'File API maps display_icon_label');

console.log('File icon display verification passed: FileItem, FileList table, grid, gallery, and API mapping all prioritize display_icon.');
