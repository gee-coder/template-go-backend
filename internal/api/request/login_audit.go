package request

// LoginAuditListQuery describes the login audit list query.
type LoginAuditListQuery struct {
	Keyword   string `form:"keyword"`
	Status    string `form:"status"`
	LoginType string `form:"loginType"`
}
