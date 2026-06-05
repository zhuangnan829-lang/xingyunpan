package service

import "testing"

func TestParseBrowserAppsJSONNormalizesExtensions(t *testing.T) {
	raw := `[{"id":1,"name":"组","items":[{"id":99,"icon":" APP ","name":"Custom Editor","extensions":".TXT, txt , .Md ","platform":"ALL","enabled":true,"max_size":0,"max_size_unit":"xx"}]}]`

	groups, err := parseBrowserAppsJSON(raw, false)
	if err != nil {
		t.Fatalf("parse browser apps failed: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Items) != 1 {
		t.Fatalf("unexpected groups: %#v", groups)
	}

	item := groups[0].Items[0]
	if item.Extensions != "txt,md" {
		t.Fatalf("unexpected normalized extensions: %s", item.Extensions)
	}
	if item.Platform != BrowserAppPlatformAll {
		t.Fatalf("unexpected platform: %s", item.Platform)
	}
	if item.MaxSize != 300 {
		t.Fatalf("unexpected max size: %d", item.MaxSize)
	}
	if item.MaxSizeUnit != "MB" {
		t.Fatalf("unexpected max size unit: %s", item.MaxSizeUnit)
	}
}

func TestResolveBrowserAppFromGroups(t *testing.T) {
	groups := []BrowserAppGroupPayload{
		{
			ID:   1,
			Name: "默认组",
			Items: []BrowserAppItemPayload{
				{ID: 10, Name: "Monaco 代码编辑器", Extensions: "txt,md,json", Platform: BrowserAppPlatformAll, Enabled: true, OpenInNewWindow: false, MaxSize: 25, MaxSizeUnit: "MB"},
			},
		},
	}

	resolved := ResolveBrowserAppFromGroups(groups, "readme.md", "text/markdown", BrowserAppPlatformAll)
	if resolved == nil {
		t.Fatal("expected browser app to be resolved")
	}
	if resolved.ID != 10 {
		t.Fatalf("unexpected resolved id: %d", resolved.ID)
	}
	if resolved.MatchedExtension != "md" {
		t.Fatalf("unexpected matched extension: %s", resolved.MatchedExtension)
	}
}

func TestParseBrowserAppsJSONRestoresBrokenPresetText(t *testing.T) {
	raw := `[{"id":1,"name":"App Group #1","description":"???","items":[{"id":3,"icon":"???","type":"???","name":"Markdown Studio","extensions":"md","create_mapping":"???","enabled":true,"max_size":20,"max_size_unit":"MB"}]}]`

	groups, err := parseBrowserAppsJSON(raw, false)
	if err != nil {
		t.Fatalf("parse browser apps failed: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Items) != 1 {
		t.Fatalf("unexpected groups: %#v", groups)
	}

	group := groups[0]
	if group.Name != "应用分组 #1" {
		t.Fatalf("unexpected group name: %s", group.Name)
	}
	item := group.Items[0]
	if item.Name != "Markdown 编辑器" {
		t.Fatalf("unexpected item name: %s", item.Name)
	}
	if item.Type != "内置应用" {
		t.Fatalf("unexpected item type: %s", item.Type)
	}
	if item.Icon != "MD" {
		t.Fatalf("unexpected item icon: %s", item.Icon)
	}
}
