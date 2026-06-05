package service

import (
	"path/filepath"
	"strings"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

func TestUserGroupListCountsMembersFromUsersTable(t *testing.T) {
	env := newUserGroupMembersEnv(t)

	payloads, err := env.userGroupService.List()
	if err != nil {
		t.Fatalf("list user groups: %v", err)
	}
	counts := userGroupCountsByName(payloads)

	if counts["Empty"] != 0 {
		t.Fatalf("empty group count = %d, want 0", counts["Empty"])
	}
	if counts["Single"] != 1 {
		t.Fatalf("single group count = %d, want 1", counts["Single"])
	}
	if counts["Many"] != 3 {
		t.Fatalf("many group count = %d, want 3", counts["Many"])
	}

	members, err := env.userGroupService.ListMembers(env.groups["Many"].ID)
	if err != nil {
		t.Fatalf("list many group members: %v", err)
	}
	if len(members) != 3 {
		t.Fatalf("many group members len = %d, want 3", len(members))
	}
	member := members[0]
	if member.Username == "" || member.Email == "" || member.Role == "" || member.UserGroupID != env.groups["Many"].ID {
		t.Fatalf("member payload missed required identity fields: %#v", member)
	}
	if member.Capacity == 0 || member.UsedSize == 0 {
		t.Fatalf("member payload missed capacity/used_size: %#v", member)
	}
}

func TestBatchMoveUserGroupRefreshesMemberCounts(t *testing.T) {
	env := newUserGroupMembersEnv(t)
	sourceID := env.groups["Many"].ID
	targetID := env.groups["Single"].ID

	moved, err := env.adminUserService.BatchUpdateGroup(&AdminUserBatchGroupPayload{
		IDs:         []uint{env.users["many-1"].ID, env.users["many-2"].ID},
		UserGroupID: targetID,
	}, 99)
	if err != nil {
		t.Fatalf("batch update group: %v", err)
	}
	if len(moved) != 2 {
		t.Fatalf("moved len = %d, want 2", len(moved))
	}

	payloads, err := env.userGroupService.List()
	if err != nil {
		t.Fatalf("list user groups after move: %v", err)
	}
	counts := userGroupCountsByName(payloads)
	if counts["Many"] != 1 {
		t.Fatalf("source group count after move = %d, want 1", counts["Many"])
	}
	if counts["Single"] != 3 {
		t.Fatalf("target group count after move = %d, want 3", counts["Single"])
	}

	sourceMembers, err := env.userGroupService.ListMembers(sourceID)
	if err != nil {
		t.Fatalf("list source members after move: %v", err)
	}
	targetMembers, err := env.userGroupService.ListMembers(targetID)
	if err != nil {
		t.Fatalf("list target members after move: %v", err)
	}
	if len(sourceMembers) != 1 || len(targetMembers) != 3 {
		t.Fatalf("member lists after move mismatch: source=%d target=%d", len(sourceMembers), len(targetMembers))
	}
}

func TestUserGroupCreateAndUpdatePersistEditableFields(t *testing.T) {
	env := newUserGroupMembersEnv(t)

	policyA := model.StoragePolicy{Name: "Create policy", Type: "local"}
	policyB := model.StoragePolicy{Name: "Update policy", Type: "local"}
	if err := env.db.Create(&policyA).Error; err != nil {
		t.Fatalf("seed create policy: %v", err)
	}
	if err := env.db.Create(&policyB).Error; err != nil {
		t.Fatalf("seed update policy: %v", err)
	}

	created, err := env.userGroupService.Create(&UserGroupPayload{
		Name:            "Regression Group",
		Description:     "created description",
		StoragePolicyID: policyA.ID,
		MaxCapacity:     quotaTwoGB,
	})
	if err != nil {
		t.Fatalf("create user group: %v", err)
	}
	if created.ID == 0 || created.Name != "Regression Group" || created.Description != "created description" ||
		created.StoragePolicyID != policyA.ID || created.StoragePolicyName != policyA.Name || created.MaxCapacity != quotaTwoGB {
		t.Fatalf("created payload did not persist fields: %#v", created)
	}

	updated, err := env.userGroupService.Update(created.ID, &UserGroupPayload{
		Name:            "Regression Group Updated",
		Description:     "updated description",
		StoragePolicyID: policyB.ID,
		MaxCapacity:     0,
	})
	if err != nil {
		t.Fatalf("update user group: %v", err)
	}
	if updated.Name != "Regression Group Updated" || updated.Description != "updated description" ||
		updated.StoragePolicyID != policyB.ID || updated.StoragePolicyName != policyB.Name || updated.MaxCapacity != 0 {
		t.Fatalf("updated payload did not persist fields: %#v", updated)
	}

	list, err := env.userGroupService.List()
	if err != nil {
		t.Fatalf("list user groups after update: %v", err)
	}
	var reloaded *UserGroupPayload
	for i := range list {
		if list[i].ID == created.ID {
			reloaded = &list[i]
			break
		}
	}
	if reloaded == nil {
		t.Fatalf("created group missing after list refresh")
	}
	if reloaded.Name != updated.Name || reloaded.Description != updated.Description ||
		reloaded.StoragePolicyID != policyB.ID || reloaded.MaxCapacity != 0 {
		t.Fatalf("reloaded payload mismatch: %#v", reloaded)
	}

	var row model.UserGroup
	if err := env.db.First(&row, created.ID).Error; err != nil {
		t.Fatalf("load user group row: %v", err)
	}
	if row.Name != updated.Name || row.Description != updated.Description ||
		row.StoragePolicyID != policyB.ID || row.MaxCapacity != 0 {
		t.Fatalf("database row mismatch: %#v", row)
	}
}

func TestUserGroupSummaryHighCapacityAndUnlimitedBuckets(t *testing.T) {
	env := newUserGroupMembersEnv(t)
	gb := int64(1024 * 1024 * 1024)
	updates := map[string]int64{
		"Empty":  0,
		"Single": 199 * gb,
		"Many":   UserGroupHighCapacityThresholdBytes,
	}
	for name, capacity := range updates {
		if err := env.db.Model(&model.UserGroup{}).
			Where("id = ?", env.groups[name].ID).
			Update("max_capacity", capacity).Error; err != nil {
			t.Fatalf("update %s capacity: %v", name, err)
		}
	}
	if err := env.db.Create(&model.SiteSetting{DefaultGroup: "Many"}).Error; err != nil {
		t.Fatalf("seed default group setting: %v", err)
	}

	summary, err := env.userGroupService.Summary()
	if err != nil {
		t.Fatalf("summary user groups: %v", err)
	}
	if summary.TotalGroups != 3 {
		t.Fatalf("total_groups = %d, want 3", summary.TotalGroups)
	}
	if summary.TotalUsers != 4 {
		t.Fatalf("total_users = %d, want 4", summary.TotalUsers)
	}
	if summary.DefaultGroup != "Many" {
		t.Fatalf("default_group = %q, want Many", summary.DefaultGroup)
	}
	if summary.HighCapacityBytes != UserGroupHighCapacityThresholdBytes {
		t.Fatalf("high_capacity_bytes = %d, want %d", summary.HighCapacityBytes, UserGroupHighCapacityThresholdBytes)
	}
	if summary.HighCapacityGroups != 1 {
		t.Fatalf("high_capacity_groups = %d, want 1", summary.HighCapacityGroups)
	}
	if summary.UnlimitedGroups != 1 {
		t.Fatalf("unlimited_groups = %d, want 1", summary.UnlimitedGroups)
	}
}

func TestUserGroupDeleteGuardsAndSuccess(t *testing.T) {
	env := newUserGroupMembersEnv(t)

	if err := env.userGroupService.Delete(999999); err == nil || !strings.Contains(err.Error(), "user group not found") {
		t.Fatalf("missing group delete error = %v, want user group not found", err)
	}

	if err := env.userGroupService.Delete(env.groups["Many"].ID); err == nil || !strings.Contains(err.Error(), "still contains users") {
		t.Fatalf("member group delete error = %v, want still contains users", err)
	}

	if err := env.db.Create(&model.SiteSetting{DefaultGroup: " Empty "}).Error; err != nil {
		t.Fatalf("seed default group setting: %v", err)
	}
	if err := env.userGroupService.Delete(env.groups["Empty"].ID); err == nil || !strings.Contains(err.Error(), "current default user group") {
		t.Fatalf("default group delete error = %v, want current default user group", err)
	}

	var setting model.SiteSetting
	if err := env.db.First(&setting).Error; err != nil {
		t.Fatalf("load default group setting: %v", err)
	}
	setting.DefaultGroup = "Single"
	if err := env.db.Save(&setting).Error; err != nil {
		t.Fatalf("move default group setting: %v", err)
	}

	var policy model.StoragePolicy
	if err := env.db.First(&policy).Error; err != nil {
		t.Fatalf("load storage policy: %v", err)
	}
	referenced := model.UserGroup{Name: "Referenced", StoragePolicyID: policy.ID, MaxCapacity: 1024}
	if err := env.db.Create(&referenced).Error; err != nil {
		t.Fatalf("seed referenced group: %v", err)
	}
	policy.GroupsJSON = `["Referenced"]`
	if err := env.db.Save(&policy).Error; err != nil {
		t.Fatalf("seed storage policy group reference: %v", err)
	}
	if err := env.userGroupService.Delete(referenced.ID); err == nil || !strings.Contains(err.Error(), "storage policy") {
		t.Fatalf("storage policy referenced group delete error = %v, want storage policy reference", err)
	}

	removable := model.UserGroup{Name: "Removable", StoragePolicyID: policy.ID, MaxCapacity: 1024}
	if err := env.db.Create(&removable).Error; err != nil {
		t.Fatalf("seed removable group: %v", err)
	}
	if err := env.userGroupService.Delete(removable.ID); err != nil {
		t.Fatalf("empty group delete should succeed: %v", err)
	}
	var remaining int64
	if err := env.db.Model(&model.UserGroup{}).Where("id = ?", removable.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("count removable group after delete: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("removable group still exists after delete")
	}
}

func TestUserGroupDeleteRejectsLastGroup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "last-user-group.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}, &model.SiteSetting{}, &model.StoragePolicy{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	policy := model.StoragePolicy{Name: "Only policy", Type: "local"}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("seed policy: %v", err)
	}
	group := model.UserGroup{Name: "Only", StoragePolicyID: policy.ID}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed only group: %v", err)
	}

	service := NewUserGroupService(
		repository.NewUserGroupRepository(db),
		repository.NewUserRepository(db),
		repository.NewStoragePolicyRepository(db),
		repository.NewSiteSettingRepository(db),
	)
	if err := service.Delete(group.ID); err == nil || !strings.Contains(err.Error(), "at least one user group") {
		t.Fatalf("last group delete error = %v, want at least one user group", err)
	}
}

type userGroupMembersEnv struct {
	db               *gorm.DB
	userGroupService UserGroupService
	adminUserService AdminUserService
	groups           map[string]model.UserGroup
	users            map[string]model.User
}

func newUserGroupMembersEnv(t *testing.T) *userGroupMembersEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "user-group-members.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserGroup{},
		&model.SiteSetting{},
		&model.StoragePolicy{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	policy := model.StoragePolicy{Name: "Members policy", Type: "local"}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("seed policy: %v", err)
	}

	groups := map[string]model.UserGroup{}
	for _, name := range []string{"Empty", "Single", "Many"} {
		group := model.UserGroup{Name: name, StoragePolicyID: policy.ID, MaxCapacity: int64(len(name)) * 1024}
		if err := db.Create(&group).Error; err != nil {
			t.Fatalf("seed group %s: %v", name, err)
		}
		groups[name] = group
	}

	users := map[string]model.User{}
	seedUser := func(username string, group model.UserGroup, role string, capacity int64, usedSize int64) {
		t.Helper()
		user := model.User{
			Username:    username,
			Email:       username + "@example.com",
			Password:    "password",
			Role:        role,
			Enabled:     true,
			UserGroupID: group.ID,
			Capacity:    capacity,
			UsedSize:    usedSize,
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user %s: %v", username, err)
		}
		users[username] = user
	}
	seedUser("single-1", groups["Single"], "admin", 1024, 128)
	seedUser("many-1", groups["Many"], "user", 2048, 256)
	seedUser("many-2", groups["Many"], "user", 4096, 512)
	seedUser("many-3", groups["Many"], "admin", 8192, 1024)

	userRepo := repository.NewUserRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)

	return &userGroupMembersEnv{
		db: db,
		userGroupService: NewUserGroupService(
			userGroupRepo,
			userRepo,
			repository.NewStoragePolicyRepository(db),
			repository.NewSiteSettingRepository(db),
		),
		adminUserService: NewAdminUserService(userRepo, userGroupRepo),
		groups:           groups,
		users:            users,
	}
}

func userGroupCountsByName(groups []UserGroupPayload) map[string]int64 {
	counts := make(map[string]int64, len(groups))
	for _, group := range groups {
		counts[group.Name] = group.UserCount
	}
	return counts
}
