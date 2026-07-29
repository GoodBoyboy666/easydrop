// Package idparser 提供 uint 类型 ID 的安全解析。
package idparser

import "strconv"

// Parse 将十进制字符串解析为当前平台可表示的 uint ID。
func Parse(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if id > uint64(^uint(0)) {
		return 0, strconv.ErrRange
	}
	return uint(id), nil
}
