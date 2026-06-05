package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// StoragePolicyController handles admin storage policy endpoints.
type StoragePolicyController struct {
	service service.StoragePolicyService
}

type migrateStoragePolicyGroupsRequest struct {
	TargetPolicyID uint   `json:"target_policy_id"`
	GroupIDs       []uint `json:"group_ids"`
}

// NewStoragePolicyController creates a storage policy controller.
func NewStoragePolicyController(storagePolicyService service.StoragePolicyService) *StoragePolicyController {
	return &StoragePolicyController{service: storagePolicyService}
}

func (c *StoragePolicyController) List(ctx *gin.Context) {
	data, err := c.service.List()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	data, err := c.service.Get(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) Preview(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	data, err := c.service.Preview(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) History(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	limit := 10
	if raw := ctx.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := c.service.History(uint(id), limit)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) RecentHits(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	limit := 20
	if raw := ctx.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	data, err := c.service.RecentHits(uint(id), limit)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) AuditDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}
	auditID, err := strconv.ParseUint(ctx.Param("auditId"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy audit id")
		return
	}

	data, err := c.service.AuditDetail(uint(id), uint(auditID))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) Rollback(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}
	auditID, err := strconv.ParseUint(ctx.Param("auditId"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy audit id")
		return
	}

	data, err := c.service.Rollback(uint(id), uint(auditID), storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy rolled back", data)
}

func (c *StoragePolicyController) MigrateGroups(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	var req migrateStoragePolicyGroupsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.MigrateGroups(uint(id), req.TargetPolicyID, req.GroupIDs, storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy user groups migrated", data)
}

func (c *StoragePolicyController) RepairLegacyDefaults(ctx *gin.Context) {
	data, err := c.service.RepairLegacyDefaults(storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "legacy storage policies repaired", data)
}

func (c *StoragePolicyController) Copy(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	data, err := c.service.Copy(uint(id), storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy copied", data)
}

func (c *StoragePolicyController) Export(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	data, err := c.service.Export(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *StoragePolicyController) Import(ctx *gin.Context) {
	overwriteID := uint(0)
	if raw := ctx.Query("overwrite_id"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "invalid overwrite storage policy id")
			return
		}
		overwriteID = uint(parsed)
	}

	var req service.StoragePolicyPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Import(&req, overwriteID, storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy imported", data)
}

func (c *StoragePolicyController) Create(ctx *gin.Context) {
	var req service.StoragePolicyPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Create(&req, storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy created", data)
}

func (c *StoragePolicyController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	var req service.StoragePolicyPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(uint(id), &req, storagePolicyActor(ctx))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy saved", data)
}

func (c *StoragePolicyController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
		return
	}

	if err := c.service.Delete(uint(id), storagePolicyActor(ctx)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "storage policy deleted", gin.H{"deleted": true})
}

func storagePolicyActor(ctx *gin.Context) service.StoragePolicyActor {
	actor := service.StoragePolicyActor{}
	if id, err := GetCurrentUserID(ctx); err == nil {
		actor.ID = id
	}
	if name, err := GetCurrentUsername(ctx); err == nil {
		actor.Name = name
	}
	return actor
}
