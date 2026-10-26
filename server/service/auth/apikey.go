package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/apikey"
)

func (s *Service) CreateAPIKey(c *gin.Context) {
	var req proto.CreateAPIKeyReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -3, "invalid session")
		return
	}

	scope, err := apikey.ParseScope(req.Scope)
	if err != nil {
		rsp.ErrRsp(c, -1, "unknown scope")
		return
	}

	secret, key, err := apikey.CreateWithScope(req.Name, principal.Username, scope)
	if err != nil {
		if errors.Is(err, apikey.ErrNameTooLong) {
			rsp.ErrRsp(c, -1, "name is too long")
			return
		}

		rsp.ErrRsp(c, -2, "failed to create api key")
		return
	}

	// The secret is returned here and nowhere else. Only its digest is kept,
	// so it cannot be shown again.
	rsp.OkRspWithData(c, &proto.CreateAPIKeyRsp{
		ID:        key.ID,
		Name:      key.Name,
		CreatedAt: key.CreatedAt,
		Key:       secret,
		Scope:     scopeName(key.Scope),
	})

	log.Debugf("created api key %s for %s", key.ID, principal.Username)
}

func (s *Service) GetAPIKeys(c *gin.Context) {
	var rsp proto.Response

	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -3, "invalid session")
		return
	}

	// An administrator sees every account's keys, with the owner of each, so a
	// key can be found and revoked whoever issued it. Anyone else sees only
	// their own.
	var keys []apikey.Key
	var err error
	if principal.Role == authn.RoleAdmin {
		keys, err = apikey.ListAll()
	} else {
		keys, err = apikey.List(principal.Username)
	}
	if err != nil {
		rsp.ErrRsp(c, -2, "failed to read api keys")
		return
	}

	listed := make([]proto.APIKey, 0, len(keys))
	for _, key := range keys {
		listed = append(listed, proto.APIKey{
			ID:        key.ID,
			Name:      key.Name,
			CreatedAt: key.CreatedAt,
			Username:  key.Username,
			Scope:     scopeName(key.Scope),
		})
	}

	rsp.OkRspWithData(c, &proto.GetAPIKeysRsp{Keys: listed})
}

func (s *Service) DeleteAPIKey(c *gin.Context) {
	var rsp proto.Response

	id := c.Param("id")
	if id == "" {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	principal, ok := middleware.CurrentPrincipal(c)
	if !ok {
		rsp.ErrRsp(c, -3, "invalid session")
		return
	}

	var err error
	if principal.Role == authn.RoleAdmin {
		err = apikey.RevokeAny(id)
	} else {
		err = apikey.Revoke(id, principal.Username)
	}
	if err != nil {
		if errors.Is(err, apikey.ErrNotFound) {
			rsp.ErrRsp(c, -1, "api key not found")
			return
		}

		rsp.ErrRsp(c, -2, "failed to revoke api key")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("%s revoked api key %s", principal.Username, id)
}

// scopeName is how the API names a key's scope.
func scopeName(scope apikey.Scope) string {
	if scope == apikey.ScopeView {
		return "view"
	}
	return "full"
}
