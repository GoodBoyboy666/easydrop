package listing

import "strings"

type SQLOrderMap map[string]string

func (m SQLOrderMap) Resolve(input, defaultKey string) string {
	if len(m) == 0 {
		return ""
	}

	if sqlOrder, ok := m[normalizeOrderKey(input)]; ok {
		return sqlOrder
	}

	return m[normalizeOrderKey(defaultKey)]
}

func normalizeOrderKey(order string) string {
	return strings.ToLower(strings.TrimSpace(order))
}
