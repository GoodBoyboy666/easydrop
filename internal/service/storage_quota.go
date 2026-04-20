package service

import (
	"context"
	"log"
	"strconv"

	"easydrop/internal/consts"
)

const defaultStorageQuotaBytes int64 = 10 * 1024 * 1024 * 1024

// getDefaultStorageQuota 读取并解析全局默认存储配额（字节），失败时返回业务错误。
func getDefaultStorageQuota(ctx context.Context, settings SettingService) (int64, error) {
	// 未注入配置服务时，直接使用内置默认值。
	if settings == nil {
		return defaultStorageQuotaBytes, nil
	}

	// 读取配置项并处理读取失败。
	quotaStr, found, err := settings.GetValue(ctx, consts.StorageQuotaSettingKey)
	if err != nil {
		log.Printf("获取全局存储配额失败: %v", err)
		return 0, ErrFailedToCalculateQuota
	}

	if !found || quotaStr == "" {
		return defaultStorageQuotaBytes, nil
	}

	// 解析并校验配置值，非法或非正数都回退到默认值。
	quota, err := strconv.ParseInt(quotaStr, 10, 64)
	if err != nil {
		log.Printf("解析存储配额失败: %v", err)
		return 0, ErrFailedToCalculateQuota
	}

	if quota <= 0 {
		return defaultStorageQuotaBytes, nil
	}

	return quota, nil
}
