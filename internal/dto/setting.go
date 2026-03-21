package dto

type SettingItem struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Desc      string `json:"desc"`
	Category  string `json:"category"`
	Sensitive bool   `json:"sensitive"`
	Public    bool   `json:"public"`
}

type SettingListInput struct {
	Category string `json:"category"`
	Key      string `json:"key"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Order    string `json:"order"`
}

type SettingListResult struct {
	Items []SettingItem `json:"items"`
	Total int64         `json:"total"`
}

type SettingUpdateInput struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

type SettingPublicItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SettingPublicResult struct {
	Items []SettingPublicItem `json:"items"`
}
