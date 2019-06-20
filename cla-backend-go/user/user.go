package user

type CLAUser struct {
	UserID string
	Name   string

	Emails []string

	LFEmail    string
	LFUsername string

	LfidProvider   UserProvider
	GithubProvider UserProvider

	ProjectIDs []string
	ClaIDs     []string
	CompanyIDs []string
}

type UserProvider struct {
	ProviderUserID string
}
