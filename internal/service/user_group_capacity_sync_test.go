package service

import (
	"strings"
	"testing"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

func TestUserGroupCapacityUpdateDoesNotSyncMembersByDefault(t *testing.T) {
	env := newUserGroupCapacitySyncEnv(t)
	user := env.user

	if _, err := env.userGroupService.Update(env.group.ID, &UserGroupPayload{
		Name:            env.group.Name,
		Description:     env.group.Description,
		StoragePolicyID: env.policy.ID,
		MaxCapacity:     quotaTwoGB,
	}); err != nil {
		t.Fatalf("update user group without sync: %v", err)
	}

	reloaded := loadUserForCapacitySync(t, env, user.ID)
	if reloaded.Capacity != quotaFiftyGB {
		t.Fatalf("member capacity changed without sync: got %d want %d", reloaded.Capacity, quotaFiftyGB)
	}
}

func TestUserGroupCapacitySyncUpdatesMembersAndBlocksFutureUpload(t *testing.T) {
	env := newUserGroupCapacitySyncEnv(t)
	user := env.user
	overTwoGB := quotaTwoGB + 1024
	if err := env.db.Model(&model.User{}).Where("id = ?", user.ID).Update("used_size", overTwoGB).Error; err != nil {
		t.Fatalf("set used_size over new capacity: %v", err)
	}

	if _, err := env.userGroupService.Update(env.group.ID, &UserGroupPayload{
		Name:               env.group.Name,
		Description:        env.group.Description,
		StoragePolicyID:    env.policy.ID,
		MaxCapacity:        quotaTwoGB,
		SyncMemberCapacity: true,
	}); err != nil {
		t.Fatalf("update user group with sync: %v", err)
	}

	reloaded := loadUserForCapacitySync(t, env, user.ID)
	if reloaded.Capacity != quotaTwoGB {
		t.Fatalf("member capacity after sync = %d, want %d", reloaded.Capacity, quotaTwoGB)
	}
	if reloaded.UsedSize != overTwoGB {
		t.Fatalf("used_size changed during sync: got %d want %d", reloaded.UsedSize, overTwoGB)
	}

	_, err := env.fileService.Upload(user.ID, "blocked.txt", 1, strings.NewReader("x"), nil)
	if err == nil {
		t.Fatal("expected upload to be blocked after synced capacity is below used_size")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("upload error = %v, want quota exceeded", err)
	}
	assertUserUsage(t, env.db, user.ID, overTwoGB)
	assertUserFileCount(t, env.db, user.ID, 0)
}

const quotaFiftyGB = int64(50 * 1024 * 1024 * 1024)

type userGroupCapacitySyncEnv struct {
	*quotaTestEnv
	policy           model.StoragePolicy
	group            model.UserGroup
	userGroupService UserGroupService
}

func newUserGroupCapacitySyncEnv(t *testing.T) *userGroupCapacitySyncEnv {
	t.Helper()

	env := newQuotaTestEnv(t, quotaFiftyGB)
	policy := model.StoragePolicy{Name: "Capacity sync policy", Type: "local"}
	if err := env.db.Create(&policy).Error; err != nil {
		t.Fatalf("seed storage policy: %v", err)
	}
	var group model.UserGroup
	if err := env.db.Where("id = ?", env.user.UserGroupID).First(&group).Error; err != nil {
		t.Fatalf("load quota group: %v", err)
	}
	group.StoragePolicyID = policy.ID
	group.MaxCapacity = quotaFiftyGB
	if err := env.db.Save(&group).Error; err != nil {
		t.Fatalf("update quota group: %v", err)
	}
	env.user.Capacity = quotaFiftyGB
	if err := env.db.Model(&model.User{}).Where("id = ?", env.user.ID).Update("capacity", quotaFiftyGB).Error; err != nil {
		t.Fatalf("set user capacity: %v", err)
	}

	return &userGroupCapacitySyncEnv{
		quotaTestEnv: env,
		policy:       policy,
		group:        group,
		userGroupService: NewUserGroupService(
			repository.NewUserGroupRepository(env.db),
			repository.NewUserRepository(env.db),
			repository.NewStoragePolicyRepository(env.db),
			repository.NewSiteSettingRepository(env.db),
		),
	}
}

func loadUserForCapacitySync(t *testing.T, env *userGroupCapacitySyncEnv, userID uint) model.User {
	t.Helper()
	var user model.User
	if err := env.db.First(&user, userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	return user
}
