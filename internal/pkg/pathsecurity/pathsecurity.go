package pathsecurity

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrBaseDirEmpty      = errors.New("基础目录不能为空")
	ErrBaseDirNotFound   = errors.New("基础目录不存在")
	ErrPathEmpty         = errors.New("路径不能为空")
	ErrContainsNUL       = errors.New("路径不能包含 NUL 字符")
	ErrAbsolutePath      = errors.New("不允许使用绝对路径")
	ErrParentDirNotFound = errors.New("目标父目录不存在")
	ErrPathTraversal     = errors.New("检测到路径穿越风险")
	ErrSymlinkNotAllowed = errors.New("路径中包含符号链接，已拒绝")
)

func SecureJoin(baseDir, userPath string) (string, error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", ErrBaseDirEmpty
	}
	if strings.TrimSpace(userPath) == "" {
		return "", ErrPathEmpty
	}
	if containsNUL(baseDir) || containsNUL(userPath) {
		return "", ErrContainsNUL
	}

	cleanInput := filepath.Clean(userPath)
	if filepath.IsAbs(cleanInput) || filepath.VolumeName(cleanInput) != "" {
		return "", ErrAbsolutePath
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(baseAbs); err != nil {
		if os.IsNotExist(err) {
			return "", ErrBaseDirNotFound
		}
		return "", err
	}

	targetAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanInput))
	if err != nil {
		return "", err
	}

	ok, err := IsPathUnderBase(baseAbs, targetAbs)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrPathTraversal
	}

	if err := checkExistingPathNoSymlink(baseAbs, ErrBaseDirNotFound); err != nil {
		return "", err
	}

	parentDir := filepath.Dir(targetAbs)
	if _, err := os.Stat(parentDir); err != nil {
		if os.IsNotExist(err) {
			return "", ErrParentDirNotFound
		}
		return "", err
	}
	if err := checkExistingPathNoSymlink(parentDir, ErrParentDirNotFound); err != nil {
		return "", err
	}

	return targetAbs, nil
}

func IsPathUnderBase(baseDir, targetPath string) (bool, error) {
	if strings.TrimSpace(baseDir) == "" {
		return false, ErrBaseDirEmpty
	}
	if strings.TrimSpace(targetPath) == "" {
		return false, ErrPathEmpty
	}
	if containsNUL(baseDir) || containsNUL(targetPath) {
		return false, ErrContainsNUL
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

func checkExistingPathNoSymlink(path string, notFoundErr error) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	cleanPath := filepath.Clean(absPath)

	for _, current := range existingPathChain(cleanPath) {
		info, statErr := os.Lstat(current)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return notFoundErr
			}
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrSymlinkNotAllowed
		}
	}

	return nil
}

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
