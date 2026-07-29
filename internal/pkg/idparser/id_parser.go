// Package idparser 提供 uint 类型 ID 的安全解析。
package idparser

import "strconv"

// Parse 将十进制字符串解析为当前平台可表示的 uint ID。
func Parse(value string) (uint, error) {
	// bitSize 为 0 时，ParseUint 会按当前平台的 uint 位宽检查范围。
	id, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
