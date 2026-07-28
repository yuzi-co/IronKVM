package proto

type LoginReq struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}

type GetAccountRsp struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ChangePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password" validate:"required"`
}

type IsPasswordUpdatedRsp struct {
	IsUpdated bool `json:"isUpdated"`
}

type UserInfo struct {
	Username      string `json:"username"`
	Role          string `json:"role"`
	Enabled       bool   `json:"enabled"`
	SystemAccount bool   `json:"systemAccount,omitempty"`
}

type ListUsersRsp struct {
	Users []UserInfo `json:"users"`
}

type CreateUserReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type UpdateUserReq struct {
	Role    *string `json:"role"`
	Enabled *bool   `json:"enabled"`
}

type CreateAPIKeyReq struct {
	Name string `json:"name"`
}

// APIKey describes an issued key. The secret is deliberately absent: it is
// only ever returned by CreateAPIKeyRsp, at the moment it is issued.
type APIKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
}

type CreateAPIKeyRsp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	// Key is shown once and cannot be recovered afterwards.
	Key string `json:"key"`
}

type GetAPIKeysRsp struct {
	Keys []APIKey `json:"keys"`
}
