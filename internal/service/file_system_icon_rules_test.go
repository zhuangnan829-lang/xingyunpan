package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeFileIconRulesAcceptsChineseCommaSpacesAndSemicolon(t *testing.T) {
	got, err := normalizeFileIconRulesJSON(`[{"label":" 图片 ","icon":" 🖼️ ","match":" .JPG， png;GIF image/* ","tint":"#abc"}]`)
	if err != nil {
		t.Fatalf("normalizeFileIconRulesJSON returned error: %v", err)
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(got), &rules); err != nil {
		t.Fatalf("normalized rules are invalid json: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(rules))
	}
	if rules[0].Label != "图片" || strings.TrimSpace(rules[0].Icon) != "🖼️" || rules[0].Match != "jpg,png,gif,image/*" || rules[0].Tint != "#abc" {
		t.Fatalf("normalized rule = %#v", rules[0])
	}
}

func TestNormalizeFileIconRulesRejectsInvalidTint(t *testing.T) {
	_, err := normalizeFileIconRulesJSON(`[{"label":"PDF","icon":"PDF","match":"pdf","tint":"red"}]`)
	if err == nil {
		t.Fatal("normalizeFileIconRulesJSON returned nil error for invalid tint")
	}
	if !strings.Contains(err.Error(), "tint") {
		t.Fatalf("error = %q, want tint error", err.Error())
	}
}

func TestNormalizeFileIconRulesRejectsNonArrayJSON(t *testing.T) {
	_, err := normalizeFileIconRulesJSON(`{"label":"PDF","icon":"PDF","match":"pdf","tint":"#fff"}`)
	if err == nil {
		t.Fatal("normalizeFileIconRulesJSON returned nil error for object json")
	}
	if err.Error() != "file icon rules must be a json array" {
		t.Fatalf("error = %q, want non-array error", err.Error())
	}
}

func TestNormalizeFileIconRulesDeduplicatesMatchTokens(t *testing.T) {
	got, err := normalizeFileIconRulesJSON(`[{"label":"PDF","icon":"PDF","match":".PDF,pdf PDF;application/pdf|APPLICATION/PDF","tint":""}]`)
	if err != nil {
		t.Fatalf("normalizeFileIconRulesJSON returned error: %v", err)
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(got), &rules); err != nil {
		t.Fatalf("normalized rules are invalid json: %v", err)
	}
	if rules[0].Match != "pdf,application/pdf" {
		t.Fatalf("match = %q, want deduplicated tokens", rules[0].Match)
	}
}
