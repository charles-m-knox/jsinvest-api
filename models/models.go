package models

type OauthState struct {
	State    string `json:"state"`
	Code     string `json:"code"`
	Verifier string
}

type PostMutationBody struct {
	Domain string `json:"d"`
	JWT    string `json:"s"`
	Field  string `json:"f"`
	Value  string `json:"v"`
	Method string `json:"m"`
}

type UserData struct {
	UserID    string
	AppID     string
	TenantID  string
	Field     string
	Value     string
	UpdatedAt int64
}

type ProductPrice struct {
	ID                     string
	ProductID              string
	RecurringInterval      string // day, week, month or year.
	RecurringIntervalCount int64  // For example, interval=month and interval_count=3 bills every 3 months.
	Price                  int64
	PriceDecimal           float64
	Currency               string
	Description            string
}

type ProductSummary struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
	Prices      []ProductPrice
}
