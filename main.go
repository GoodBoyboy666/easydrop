package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"easydrop/internal/di"
	"easydrop/internal/router"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// @title           EasyDrop API
// @version         1.0
// @description     这是一个轻量级的说说服务
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	var configDir string
	flag.StringVar(&configDir, "config-dir", "", "config directory containing config.yaml")
	flag.Parse()
	args := flag.Args()

	if len(args) > 0 && args[0] == "generate-jwt-token" {
		outputDir, forceOverwrite, err := parseGenerateJWTTokenArgs(args[1:])
		if err != nil {
			log.Fatal(err)
		}

		if err := generateJWTTokenFiles(outputDir, forceOverwrite); err != nil {
			log.Fatalf("生成 JWT 密钥文件失败: %v", err)
		}
		log.Printf("JWT 密钥文件已生成: %s, %s", filepath.Join(outputDir, "private.pem"), filepath.Join(outputDir, "public.pem"))
		return
	}

	if configDir == "" {
		configDir = "data"
	}

	app, err := di.Initialize(configDir, false)
	if err != nil {
		log.Fatalf("初始化应用失败: %v\n", err)
	}
	if app == nil || app.Config == nil {
		log.Fatal("初始化应用失败: 应用配置为空")
	}

	engine := router.BuildEngine(app)

	serverCfg := app.Config.Server
	addr := strings.TrimSpace(serverCfg.Addr)
	if addr == "" {
		addr = ":8080"
	}
	readTimeout := serverCfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 10 * time.Second
	}
	writeTimeout := serverCfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 15 * time.Second
	}
	shutdownTimeout := serverCfg.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP 服务启动中，监听地址: %s", addr)
		if runErr := srv.ListenAndServe(); runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
			errCh <- runErr
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case runErr := <-errCh:
		log.Fatalf("HTTP 服务启动失败: %v", runErr)
	case sig := <-sigCh:
		log.Printf("收到退出信号: %s，开始关闭", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP 服务关闭失败: %v", err)
	}

	log.Println("HTTP 服务已关闭")
}

func parseGenerateJWTTokenArgs(args []string) (string, bool, error) {
	outputDir := "."
	forceOverwrite := false

	for _, arg := range args {
		switch arg {
		case "--force":
			forceOverwrite = true
		default:
			if strings.HasPrefix(arg, "-") {
				return "", false, fmt.Errorf("未知参数: %s，用法: generate-jwt-token [目录路径] [--force]", arg)
			}
			if outputDir != "." {
				return "", false, errors.New("用法: generate-jwt-token [目录路径] [--force]")
			}
			outputDir = arg
		}
	}

	return outputDir, forceOverwrite, nil
}

func generateJWTTokenFiles(outputDir string, forceOverwrite bool) error {
	dir := strings.TrimSpace(outputDir)
	if dir == "" {
		dir = "."
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成 RSA 私钥失败: %w", err)
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("编码 RSA 公钥失败: %w", err)
	}

	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")

	if err := writePEMFile(privatePath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey), 0o600, forceOverwrite); err != nil {
		return err
	}
	if err := writePEMFile(publicPath, "PUBLIC KEY", publicDER, 0o644, forceOverwrite); err != nil {
		return err
	}

	return nil
}

func writePEMFile(path string, blockType string, content []byte, perm os.FileMode, forceOverwrite bool) error {
	flag := os.O_WRONLY | os.O_CREATE
	if forceOverwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}

	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("文件已存在: %s（如需覆盖请加 --force）", path)
		}
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: content}); err != nil {
		return fmt.Errorf("写入 PEM 文件失败: %w", err)
	}

	return nil
}
