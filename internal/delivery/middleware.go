package delivery

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reservista.kz/internal/domain"
)

const (
	authorizationHeader = "Authorization"

	idCtx   = "userId"
	roleCtx = "userRoles"
)

func (h *Handler) userIdentity(c *gin.Context) {
	id, roles, err := h.parseAuthHeader(c)
	if err != nil {
		switch err.Error() {
		case domain.ErrTokenExpired.Error():
			h.refresh(c)
			c.Next()
		case http.ErrNoCookie.Error(), domain.ErrUnauthorized.Error(), domain.ErrTokenInvalidElements.Error():
			newResponse(c, http.StatusUnauthorized, "unauthorized access: "+err.Error())
			return
		case domain.ErrTokenExpired.Error():
			break
		default:
			newResponse(c, http.StatusInternalServerError, "failed to parse jwt to id: "+err.Error())
			return
		}
	}
	c.Set(idCtx, id)
	c.Set(roleCtx, roles)
	c.Next()
}
func (h *Handler) parseAuthHeader(c *gin.Context) (string, []string, error) {
	token, err := c.Cookie("jwt")
	if err != nil {
		return "", nil, err
	}
	id, roles, err := h.TokenManager.Parse(token)
	if err != nil {
		return "", nil, err
	}
	return id, roles, nil
}

func (h *Handler) isPermitted(permittedRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get(roleCtx)

		if !exists {
			newResponse(c, http.StatusUnauthorized, "unauthorized access: missing roles")
			return
		}

		if !hasAnyPermittedRole(userRoles.([]string), permittedRoles) {
			newResponse(c, http.StatusUnauthorized, "unauthorized access: access denied due to RBAC")
		}
		c.Next()
	}
}

// hasAnyPermittedRole checks if there's any intersection between userRoles and permittedRoles.
func hasAnyPermittedRole(userRoles []string, permittedRoles []string) bool {
	permittedSet := make(map[string]bool)
	for _, role := range permittedRoles {
		permittedSet[role] = true
	}

	for _, role := range userRoles {
		if permittedSet[role] {
			return true
		}
	}

	return false
}

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Credentials", "true")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		//TODO: solve problem with CORS policy
		c.AbortWithStatus(http.StatusOK)
	}
}
