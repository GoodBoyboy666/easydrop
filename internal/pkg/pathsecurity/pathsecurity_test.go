package pathsecurity

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSecureJoin(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	uploadDir := filepath.Join(baseDir, "upload")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	t.Run("正常拼接", func(t *testing.T) {
		got, err := SecureJoin(baseDir, filepath.Join("upload", "a.txt"))
		if err != nil {
			t.Fatalf("期望无错误，实际为: %v", err)
		}

		want := filepath.Join(baseDir, "upload", "a.txt")
		if got != want {
			t.Fatalf("拼接结果不符合预期，want=%s，got=%s", want, got)
		}
	})

	t.Run("路径穿越", func(t *testing.T) {
		_, err := SecureJoin(baseDir, filepath.Join("..", "secret.txt"))
		if !errors.Is(err, ErrPathTraversal) {
			t.Fatalf("期望错误 ErrPathTraversal，实际为: %v", err)
		}
	})

	t.Run("绝对路径", func(t *testing.T) {
		_, err := SecureJoin(baseDir, filepath.Join(baseDir, "x.txt"))
		if !errors.Is(err, ErrAbsolutePath) {
			t.Fatalf("期望错误 ErrAbsolutePath，实际为: %v", err)
		}
	})

	t.Run("包含NUL字符", func(t *testing.T) {
		_, err := SecureJoin(baseDir, "a\x00b.txt")
		if !errors.Is(err, ErrContainsNUL) {
			t.Fatalf("期望错误 ErrContainsNUL，实际为: %v", err)
		}
	})

	t.Run("父目录不存在", func(t *testing.T) {
		_, err := SecureJoin(baseDir, filepath.Join("not-exist", "a.txt"))
		if !errors.Is(err, ErrParentDirNotFound) {
			t.Fatalf("期望错误 ErrParentDirNotFound，实际为: %v", err)
		}
	})
}

func TestSecureJoin_RejectAnySymlink(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(baseDir, "real"), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	linkDir := filepath.Join(baseDir, "link")
	if err := os.Symlink(filepath.Join(baseDir, "real"), linkDir); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("Windows 环境可能缺少创建符号链接权限，跳过测试: %v", err)
		}
		t.Fatalf("创建符号链接失败: %v", err)
	}

	_, err := SecureJoin(baseDir, filepath.Join("link", "a.txt"))
	if !errors.Is(err, ErrSymlinkNotAllowed) {
		t.Fatalf("期望错误 ErrSymlinkNotAllowed，实际为: %v", err)
	}
}

func TestSecureJoin_BaseDirIsSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	realBase := filepath.Join(root, "real-base")
	if err := os.MkdirAll(realBase, 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	baseLink := filepath.Join(root, "base-link")
	if err := os.Symlink(realBase, baseLink); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("Windows 环境可能缺少创建符号链接权限，跳过测试: %v", err)
		}
		t.Fatalf("创建符号链接失败: %v", err)
	}

	_, err := SecureJoin(baseLink, "a.txt")
	if !errors.Is(err, ErrSymlinkNotAllowed) {
		t.Fatalf("期望错误 ErrSymlinkNotAllowed，实际为: %v", err)
	}
}

func TestIsPathUnderBase(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	inside := filepath.Join(baseDir, "a", "b.txt")
	outside := filepath.Join(baseDir, "..", "other", "x.txt")

	ok, err := IsPathUnderBase(baseDir, inside)
	if err != nil {
		t.Fatalf("期望无错误，实际为: %v", err)
	}
	if !ok {
		t.Fatalf("inside 应位于基础目录内")
	}

	ok, err = IsPathUnderBase(baseDir, outside)
	if err != nil {
		t.Fatalf("期望无错误，实际为: %v", err)
	}
	if ok {
		t.Fatalf("outside 不应位于基础目录内")
	}

	_, err = IsPathUnderBase(baseDir, "a\x00b")
	if !errors.Is(err, ErrContainsNUL) {
		t.Fatalf("期望错误 ErrContainsNUL，实际为: %v", err)
	}
}

