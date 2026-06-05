package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// NodeController handles admin node endpoints.
type NodeController struct {
	service service.NodeService
}

// NewNodeController creates a node controller.
func NewNodeController(nodeService service.NodeService) *NodeController {
	return &NodeController{service: nodeService}
}

func (c *NodeController) List(ctx *gin.Context) {
	data, err := c.service.List()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *NodeController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid node id")
		return
	}

	data, err := c.service.Get(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *NodeController) Create(ctx *gin.Context) {
	var req service.NodePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Create(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "node created", data)
}

func (c *NodeController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid node id")
		return
	}

	var req service.NodePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(uint(id), &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "node saved", data)
}

func (c *NodeController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid node id")
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "node deleted", gin.H{"deleted": true})
}

func (c *NodeController) TestOfflineConnectivity(ctx *gin.Context) {
	var req service.NodeOfflinePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := service.TestNodeOfflineConnectivity(req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "node connectivity checked", data)
}

func (c *NodeController) CheckHealth(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid node id")
		return
	}

	data, err := c.service.CheckHealth(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "node health checked", data)
}
