package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/logging"
	"harns-platform/pkg/platform/principal"
)

type loginReq struct {
	Username string `json:"username" binding:"required,max=64"`
	Password string `json:"password" binding:"required,max=512"`
	Code     string `json:"code"`
	Key      string `json:"key"`
}

type loginResp struct {
	Token string `json:"token"`
}

type captchaResp struct {
	UUID string `json:"uuid"`
	Img  string `json:"img"`
}

// Login handles POST /login.
func (s *Service) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.V(3).InfoS("Invalid login request", "err", err)
		response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
		return
	}

	token, user, err := s.login(c.Request.Context(), req.Username, req.Password, req.Code, req.Key, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		logging.V(2).InfoS("Login failed", "username", req.Username, "err", err)
		writeAuthError(c, err)
		return
	}
	if user != nil {
		ctx := principal.With(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)
	}
	response.WriteOK(c, loginResp{Token: token})
}

// VerificationCode handles POST /verificationCode.
func (s *Service) VerificationCode(c *gin.Context) {
	uuid, img, err := s.generateCaptcha()
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.WriteOK(c, captchaResp{UUID: uuid, Img: img})
}

// Logout handles POST /logout.
func (s *Service) Logout(c *gin.Context) {
	s.logout(c.Request.Context(), c.GetHeader(s.cfg.TokenHeader))
	response.WriteOK(c, nil)
}

// Index handles GET /index?refresh=.
func (s *Service) Index(c *gin.Context) {
	sess := sessionFromContext(c)
	if sess == nil {
		writeAuthError(c, response.ErrUnauthorized())
		return
	}
	refresh, _ := strconv.ParseBool(c.Query("refresh"))
	user, err := s.index(c.Request.Context(), sess, refresh)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.WriteOK(c, user)
}

// PublicKey handles GET /publicKey.
func (s *Service) PublicKey(c *gin.Context) {
	key, err := s.publicKey()
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.WriteOK(c, key)
}

func writeAuthError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	switch response.ErrorCode(err) {
	case int(response.ErrCodeLoginExpired), int(response.ErrCodeUnauthorized):
		status = http.StatusUnauthorized
	case int(response.ErrCodeLoginFailed):
		status = http.StatusInternalServerError
	}
	response.WriteError(c, status, err)
}
