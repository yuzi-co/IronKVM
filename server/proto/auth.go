package proto

type LoginReq struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}

type GetAccountRsp struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	// SystemAccount is true for the device owner, whose password change also
	// sets the root password used by SSH and the console.
	SystemAccount bool `json:"systemAccount"`
	// TOTPEnabled is true when the web login asks this account for a code.
	TOTPEnabled bool `json:"totpEnabled"`
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
	TOTPEnabled   bool   `json:"totpEnabled,omitempty"`
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
	// Username is the account the key acts as.
	Username string `json:"username"`
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

// LoginChallengeRsp is the data of a login that needs a second step: the
// password was right and the account has TOTP on. Challenge names the half
// finished login for LoginTOTPReq.
type LoginChallengeRsp struct {
	TOTPRequired bool   `json:"totpRequired"`
	Challenge    string `json:"challenge"`
}

// LoginTOTPReq is the second login step. Code is a TOTP code or one of the
// account's recovery codes.
type LoginTOTPReq struct {
	Challenge string `json:"challenge" validate:"required,max=128"`
	Code      string `json:"code" validate:"required,max=64"`
}

type GetTOTPRsp struct {
	Enabled bool `json:"enabled"`
	// RecoveryCodes is how many unused recovery codes are left.
	RecoveryCodes int `json:"recoveryCodes"`
}

// TOTPPasswordReq confirms a change to the second factor with the account's
// password, encrypted the way the login form sends it.
type TOTPPasswordReq struct {
	Password string `json:"password" validate:"required"`
}

// SetupTOTPRsp carries a new secret that is not in force yet. EnableTOTPReq
// puts it in force once a code from it has been typed back.
type SetupTOTPRsp struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type EnableTOTPReq struct {
	Code string `json:"code" validate:"required,max=16"`
}

// RecoveryCodesRsp is the only time recovery codes are shown. The account
// file keeps their hashes.
type RecoveryCodesRsp struct {
	RecoveryCodes []string `json:"recoveryCodes"`
}

// ShareLinkInfo is a share link as an administrator sees it. The token is not
// here: it is shown once, when the link is made.
type ShareLinkInfo struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

type ListShareLinksRsp struct {
	Links []ShareLinkInfo `json:"links"`
}

type CreateShareLinkReq struct {
	Label string `json:"label"`
	// TTLSeconds is how long the link works.
	TTLSeconds int64 `json:"ttlSeconds" validate:"required"`
}

type CreateShareLinkRsp struct {
	ShareLinkInfo
	// Token goes in the URL the link is shared by, and cannot be read again.
	Token string `json:"token"`
}

type RedeemShareLinkReq struct {
	Token string `json:"token" validate:"required"`
}
