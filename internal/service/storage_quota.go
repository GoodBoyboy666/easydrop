package service

import (
	"context"
	"log"
	"strconv"
)

const defaultStorageQuotaBytes int64 = 10 * 1024 * 1024 * 1024

func getDefaultStorageQuota(ctx context.Context, settings SettingService) (int64, error) {
	if settings == nil {
		return defaultStorageQuotaBytes, nil
	}

	quotaStr, found, err := settings.GetValue(ctx, "storage.quota")
	if err != nil {
		log.Printf("获取全局存储配额失败: %v", err)
		return 0, ErrFailedToCalculateQuota
	}

	if !found || quotaStr == "" {
		return defaultStorageQuotaBytes, nil
	}

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
