package main

import (
	"fmt"
	"testing"
)

func TestShouldIgnoreSettingMigrationErrorForWrappedQueueSettingsLegacyDrop(t *testing.T) {
	err := fmt.Errorf("migrate queue settings failed: Error 1091 (42000): Can't DROP 'uni_queue_settings_queue_key'; check that column/key exists")
	if !shouldIgnoreSettingMigrationError(err) {
		t.Fatalf("expected wrapped queue_settings legacy drop error to be ignored")
	}
}

func TestShouldNotIgnoreUnrelatedQueueSettingsDrop(t *testing.T) {
	err := fmt.Errorf("migrate queue settings failed: Error 1091 (42000): Can't DROP 'some_other_key'; check that column/key exists")
	if shouldIgnoreSettingMigrationError(err) {
		t.Fatalf("unexpectedly ignored unrelated drop error")
	}
}
