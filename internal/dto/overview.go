package dto

// AdminOverviewTotals 表示后台概览页的核心汇总指标。
type AdminOverviewTotals struct {
	Users       int64 `json:"users"`
	Posts       int64 `json:"posts"`
	Comments    int64 `json:"comments"`
	Attachments int64 `json:"attachments"`
}

// AdminOverviewTrendItem 表示单日趋势数据。
type AdminOverviewTrendItem struct {
	Date     string `json:"date"`
	Posts    int64  `json:"posts"`
	Comments int64  `json:"comments"`
}

// AdminOverviewResult 表示后台概览聚合结果。
type AdminOverviewResult struct {
	Totals         AdminOverviewTotals      `json:"totals"`
	RecentActivity []AdminOverviewTrendItem `json:"recent_activity"`
}
