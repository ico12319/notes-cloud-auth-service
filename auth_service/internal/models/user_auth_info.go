package models

type UserPersonalInfo struct {
	Name  string
	Email string
}

type UserOIDCInfo struct {
	ProviderUserID string
	Provider       string
}

type UserAuthInfo struct {
	UserPersonalInfo
	UserOIDCInfo
}
