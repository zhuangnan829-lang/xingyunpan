package controller

import (
	"net/http"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userService    service.UserService
	avatarService  service.AvatarService
	captchaService service.CaptchaRuntimeService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// NewUserControllerWithAvatarService creates a user controller with runtime avatar settings.
func NewUserControllerWithAvatarService(userService service.UserService, avatarService service.AvatarService) *UserController {
	return &UserController{
		userService:   userService,
		avatarService: avatarService,
	}
}

func NewUserControllerWithRuntimeServices(userService service.UserService, avatarService service.AvatarService, captchaService service.CaptchaRuntimeService) *UserController {
	return &UserController{
		userService:    userService,
		avatarService:  avatarService,
		captchaService: captchaService,
	}
}

// Register 用户注册接口
func (c *UserController) Register(ctx *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		Email         string `json:"email" binding:"required,email"`
		EmailCode     string `json:"email_code" binding:"required"`
		CaptchaToken  string `json:"captcha_token"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.verifyCaptcha(ctx, service.CaptchaSceneRegister, req.Email, req.CaptchaToken, req.CaptchaID, req.CaptchaAnswer, nil); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	user, err := c.userService.Register(req.Username, req.Password, req.Email, req.EmailCode)
	if err != nil {
		_ = c.reportCaptchaFailure(ctx, service.CaptchaSceneRegister, req.Email)
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.reportCaptchaSuccess(ctx, service.CaptchaSceneRegister, req.Email)

	response.SuccessWithMessage(ctx, "注册成功", gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// SendRegisterEmailCode 发送注册邮箱验证码
func (c *UserController) SendRegisterEmailCode(ctx *gin.Context) {
	var req struct {
		Email         string `json:"email" binding:"required,email"`
		CaptchaToken  string `json:"captcha_token"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.verifyCaptcha(ctx, service.CaptchaSceneRegister, req.Email, req.CaptchaToken, req.CaptchaID, req.CaptchaAnswer, nil); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.userService.SendRegisterEmailCode(req.Email); err != nil {
		_ = c.reportCaptchaFailure(ctx, service.CaptchaSceneRegister, req.Email)
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.reportCaptchaSuccess(ctx, service.CaptchaSceneRegister, req.Email)

	response.SuccessWithMessage(ctx, "验证码已发送，请注意查收邮箱", nil)
}

// SendResetPasswordEmailCode 发送找回密码邮箱验证码
func (c *UserController) SendResetPasswordEmailCode(ctx *gin.Context) {
	var req struct {
		Email         string `json:"email" binding:"required,email"`
		CaptchaToken  string `json:"captcha_token"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.verifyCaptcha(ctx, service.CaptchaSceneResetPassword, req.Email, req.CaptchaToken, req.CaptchaID, req.CaptchaAnswer, nil); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.userService.SendResetPasswordEmailCode(req.Email); err != nil {
		_ = c.reportCaptchaFailure(ctx, service.CaptchaSceneResetPassword, req.Email)
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.reportCaptchaSuccess(ctx, service.CaptchaSceneResetPassword, req.Email)

	response.SuccessWithMessage(ctx, "验证码已发送，请注意查收邮箱", nil)
}

// ResetPasswordByEmailCode 通过邮箱验证码重置密码
func (c *UserController) ResetPasswordByEmailCode(ctx *gin.Context) {
	var req struct {
		Email         string `json:"email" binding:"required,email"`
		EmailCode     string `json:"email_code" binding:"required"`
		NewPassword   string `json:"new_password" binding:"required"`
		CaptchaToken  string `json:"captcha_token"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.verifyCaptcha(ctx, service.CaptchaSceneResetPassword, req.Email, req.CaptchaToken, req.CaptchaID, req.CaptchaAnswer, nil); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.userService.ResetPasswordByEmailCode(req.Email, req.EmailCode, req.NewPassword); err != nil {
		_ = c.reportCaptchaFailure(ctx, service.CaptchaSceneResetPassword, req.Email)
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	_ = c.reportCaptchaSuccess(ctx, service.CaptchaSceneResetPassword, req.Email)

	response.SuccessWithMessage(ctx, "密码重置成功，请使用新密码登录", nil)
}

// Login 用户登录接口
func (c *UserController) Login(ctx *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		CaptchaToken  string `json:"captcha_token"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
		SliderTrack   []int  `json:"slider_track"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.verifyCaptcha(ctx, service.CaptchaSceneLogin, req.Username, req.CaptchaToken, req.CaptchaID, req.CaptchaAnswer, req.SliderTrack); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := c.userService.Login(req.Username, req.Password)
	if err != nil {
		_ = c.reportCaptchaFailure(ctx, service.CaptchaSceneLogin, req.Username)
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	_ = c.reportCaptchaSuccess(ctx, service.CaptchaSceneLogin, req.Username)

	response.SuccessWithMessage(ctx, "登录成功", resp)
}

// GetUserInfo 获取用户信息接口
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	user, err := c.userService.GetUserInfo(userID.(uint))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, user)
}

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	user, err := c.userService.UpdateProfile(userID.(uint), req.Username)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "profile updated", user)
}

func (c *UserController) UploadAvatar(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := userIDValue.(uint)

	if c.avatarService == nil {
		response.Error(ctx, http.StatusInternalServerError, "avatar service is unavailable")
		return
	}
	settings, err := c.avatarService.GetRuntimeSettings()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, settings.LimitBytes+(1<<20))
	file, err := ctx.FormFile("avatar")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "please choose an avatar image")
		return
	}
	avatarURL, err := c.avatarService.SaveUploadedAvatar(userID, file)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	user, err := c.userService.UpdateAvatar(userID, avatarURL)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "avatar updated", user)
}

func (c *UserController) verifyCaptcha(ctx *gin.Context, scene service.CaptchaScene, identity, token, captchaID, answer string, sliderTrack []int) error {
	if c.captchaService == nil {
		return nil
	}
	return c.captchaService.Verify(service.CaptchaInput{
		Scene:         scene,
		Path:          ctx.FullPath(),
		Identity:      identity,
		RemoteIP:      ctx.ClientIP(),
		CaptchaToken:  token,
		CaptchaID:     captchaID,
		CaptchaAnswer: answer,
		SliderTrack:   sliderTrack,
	})
}

func (c *UserController) reportCaptchaFailure(ctx *gin.Context, scene service.CaptchaScene, identity string) error {
	if c.captchaService == nil {
		return nil
	}
	return c.captchaService.ReportFailure(scene, identity, ctx.ClientIP())
}

func (c *UserController) reportCaptchaSuccess(ctx *gin.Context, scene service.CaptchaScene, identity string) error {
	if c.captchaService == nil {
		return nil
	}
	return c.captchaService.ReportSuccess(scene, identity, ctx.ClientIP())
}

func (c *UserController) ChangePassword(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := c.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "password changed", gin.H{"updated": true})
}

func (c *UserController) GetPreferences(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	preferences, err := c.userService.GetPreferences(userID.(uint))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, preferences)
}

func (c *UserController) UpdatePreferences(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.UserPreferencePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	preferences, err := c.userService.UpdatePreferences(userID.(uint), req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "preferences updated", preferences)
}
