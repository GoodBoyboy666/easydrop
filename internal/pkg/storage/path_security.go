package storage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	errBaseDirEmpty = errors.New("基础目录不能为空")
	errPathEmpty    = errors.New("路径不能为空")
	errContainsNUL  = errors.New("路径不能包含 NUL 字符")
)

// isPathUnderBase 判断 targetPath 是否位于 baseDir 下。
func isPathUnderBase(baseDir, targetPath string) (bool, error) {
	if strings.TrimSpace(baseDir) == "" {
		return false, errBaseDirEmpty
	}
	if strings.TrimSpace(targetPath) == "" {
		return false, errPathEmpty
	}
	if containsNUL(baseDir) || containsNUL(targetPath) {
		return false, errContainsNUL
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return false, err
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return false, err
	}

	baseClean := filepath.Clean(baseAbs)
	targetClean := filepath.Clean(targetAbs)

	if runtime.GOOS == "windows" {
		if !strings.EqualFold(filepath.VolumeName(baseClean), filepath.VolumeName(targetClean)) {
			return false, nil
		}
	}

	rel, err := filepath.Rel(baseClean, targetClean)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return true, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, nil
	}

	return true, nil
}

// ensureExistingPathChainNoSymlink 检查现有目录链路不包含符号链接。
func ensureExistingPathChainNoSymlink(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, current := range existingPathChain(filepath.Clean(absPath)) {
		info, statErr := os.Lstat(current)
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrInvalidLocalObjectKey
		}
	}

	return nil
}

// ensureNoSymlinkOnExistingPrefix 检查 base 到 target 的已存在前缀不包含符号链接。
func ensureNoSymlinkOnExistingPrefix(basePath, targetPath string) error {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}

	current := basePath
	for _, segment := range strings.Split(rel, string(filepath.Separator)) {
		if segment == "" || segment == "." {
			continue
		}

		current = filepath.Join(current, segment)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				return nil
			}
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrInvalidLocalObjectKey
		}
	}

	return nil
}

// existingPathChain 返回从根到目标路径的所有已拼接链路片段。
func existingPathChain(path string) []string {
	volume := filepath.VolumeName(path)
	rest := strings.TrimPrefix(path, volume)
	parts := strings.Split(rest, string(os.PathSeparator))

	chain := make([]string, 0, len(parts)+1)
	current := volume + string(os.PathSeparator)
	if volume == "" {
		current = string(os.PathSeparator)
	}
	chain = append(chain, current)

	for _, p := range parts {
		if p == "" {
			continue
		}
		current = filepath.Join(current, p)
		chain = append(chain, current)
	}

	return chain
}

func containsNUL(s string) bool {
	return strings.ContainsRune(s, '\x00')
}
