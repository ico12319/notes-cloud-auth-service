package models

type UserPersonalInfo struct {
	FirstName string
	LastName  string
	Email     string
}

type UserOIDCInfo struct {
	ProviderUserID string
	Provider       string
}

type UserAuthInfo struct {
	UserPersonalInfo
	UserOIDCInfo
}
