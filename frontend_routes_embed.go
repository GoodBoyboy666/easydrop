//go:build embed_frontend

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var frontendFiles embed.FS

var frontendDistFS = mustFrontendDistFS()
var frontendIndexData = mustReadFrontendFile("index.html")

// registerFrontendRoutes 注册内嵌前端的 SPA 回退逻辑。
func registerFrontendRoutes(r *gin.Engine) {
	if r == nil {
		return
	}

	r.NoRoute(func(c *gin.Context) {
		requestPath := path.Clean("/" + strings.TrimPrefix(c.Request.URL.Path, "/"))
		if requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		filePath := strings.TrimPrefix(requestPath, "/")
		if filePath == "" || filePath == "." {
			c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexData)
			return
		}

		file, err := frontendDistFS.Open(filePath)
		if err == nil {
			defer func() { _ = file.Close() }()

			info, statErr := file.Stat()
			if statErr == nil && !info.IsDir() {
				c.FileFromFS(filePath, http.FS(frontendDistFS))
				return
			}
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexData)
	})
}

// mustFrontendDistFS 返回裁剪到 web/dist 的嵌入文件系统，初始化失败时直接终止启动。
func mustFrontendDistFS() fs.FS {
	distFS, err := fs.Sub(frontendFiles, "web/dist")
	if err != nil {
		panic("初始化内嵌前端文件系统失败: " + err.Error())
	}
	return distFS
}

// mustReadFrontendFile 读取内嵌前端中的固定文件，初始化失败时直接终止启动。
func mustReadFrontendFile(name string) []byte {
	content, err := fs.ReadFile(frontendDistFS, name)
	if err != nil {
		panic("读取内嵌前端文件失败: " + err.Error())
	}

	return content
}
