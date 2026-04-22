package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"easydrop/internal/config"
	"easydrop/internal/di"
	"easydrop/internal/router"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

const generateJWTTokenCommand = "generate-jwt-token"

var (
	appDisplayName = "EasyDrop"
	appVersion     = "dev"
	buildTime      = "unknown"
	gitCommit      = "unknown"
)

// @title           EasyDrop API
// @version         1.0
// @description     这是一个轻量级的说说服务
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	if err := newRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	var configDir string
	var autoGenerateJWT bool

	cmd := &cobra.Command{
		Use:           "easydrop",
		Short:         "EasyDrop 服务端程序",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: strings.Join([]string{
			"easydrop",
			"easydrop --config-dir data",
			"easydrop --auto-generate-jwt",
			"easydrop --config-dir /etc/easydrop --auto-generate-jwt",
			"easydrop generate-jwt-token data/jwt --force",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(configDir, autoGenerateJWT)
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().StringVar(&configDir, "config-dir", "", "config directory containing config.yaml")
	cmd.Flags().BoolVar(&autoGenerateJWT, "auto-generate-jwt", false, "启动时自动检查 JWT 密钥，缺失时生成")
	cmd.AddCommand(newGenerateJWTTokenCommand())

	return cmd
}

func newGenerateJWTTokenCommand() *cobra.Command {
	var forceOverwrite bool

	cmd := &cobra.Command{
		Use:   generateJWTTokenCommand + " [目录路径]",
		Short: "生成 JWT 私钥和公钥文件",
		Long: strings.Join([]string{
			"生成 JWT 私钥和公钥文件，默认输出到当前目录。",
			"输出文件名固定为 private.pem 和 public.pem。",
		}, "\n"),
		Args: cobra.MaximumNArgs(1),
		Example: strings.Join([]string{
			"easydrop generate-jwt-token",
			"easydrop generate-jwt-token data/jwt",
			"easydrop generate-jwt-token data/jwt --force",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) == 1 {
				outputDir = args[0]
			}

			if err := generateJWTTokenFiles(outputDir, forceOverwrite); err != nil {
				return fmt.Errorf("生成 JWT 密钥文件失败: %w", err)
			}

			log.Printf("JWT 密钥文件已生成: %s, %s", filepath.Join(outputDir, "private.pem"), filepath.Join(outputDir, "public.pem"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&forceOverwrite, "force", false, "已存在时覆盖 private.pem 和 public.pem")

	return cmd
}

func runServer(configDir string, autoGenerateJWT bool) error {
	if configDir == "" {
		configDir = "data"
	}

	printBuildInfoBanner(os.Stdout)

	if err := ensureDefaultConfigOnStartup(configDir, log.Default()); err != nil {
		return fmt.Errorf("启动前检查配置文件失败: %w", err)
	}

	if err := ensureJWTKeysOnStartup(configDir, autoGenerateJWT, log.Default()); err != nil {
		return fmt.Errorf("启动前检查 JWT 密钥失败: %w", err)
	}

	app, err := di.Initialize(configDir, false)
	if err != nil {
		return fmt.Errorf("初始化应用失败: %w", err)
	}
	if app == nil || app.Config == nil {
		return errors.New("初始化应用失败: 应用配置为空")
	}
	if err := prepareInitSecret(context.Background(), app, log.Default()); err != nil {
		return err
	}

	engine := router.BuildEngine(app)
	registerFrontendRoutes(engine)

	serverCfg := app.Config.Server
	addr := serverCfg.Addr
	readTimeout := serverCfg.ReadTimeout
	writeTimeout := serverCfg.WriteTimeout
	shutdownTimeout := serverCfg.ShutdownTimeout

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
		return fmt.Errorf("HTTP 服务启动失败: %w", runErr)
	case sig := <-sigCh:
		log.Printf("收到退出信号: %s，开始关闭", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP 服务关闭失败: %w", err)
	}

	log.Println("HTTP 服务已关闭")
	return nil
}

func prepareInitSecret(ctx context.Context, app *di.App, logger *log.Logger) error {
	if app == nil {
		return errors.New("初始化保护失败: 应用为空")
	}
	if app.InitService == nil || app.InitSecretGuard == nil {
		return errors.New("初始化保护失败: 依赖未正确初始化")
	}
	if logger == nil {
		logger = log.Default()
	}

	status, err := app.InitService.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("读取系统初始化状态失败: %w", err)
	}
	if status == nil || status.Initialized {
		return nil
	}

	secret, err := app.InitSecretGuard.EnsureSecret(ctx)
	if err != nil {
		return fmt.Errorf("生成 init secret 失败: %w", err)
	}

	logger.Printf("系统未初始化，请在初始化请求中提交 init secret: %s", secret)
	return nil
}

func printBuildInfoBanner(w io.Writer) {
	if w == nil {
		return
	}

	title, err := pterm.DefaultBigText.
		WithWriter(w).
		WithLetters(
			putils.LettersFromStringWithStyle(buildValueOrDefault(appDisplayName, "EasyDrop"), pterm.NewStyle(pterm.FgLightCyan)),
		).
		Srender()
	if err == nil {
		fmt.Fprintln(w, title)
	}

	content := strings.Join([]string{
		fmt.Sprintf("Program    : %s", buildValueOrDefault(appDisplayName, "EasyDrop")),
		fmt.Sprintf("Version    : %s", buildValueOrDefault(appVersion, "dev")),
		fmt.Sprintf("Build Time : %s", buildValueOrDefault(buildTime, "unknown")),
		fmt.Sprintf("Commit     : %s", buildValueOrDefault(gitCommit, "unknown")),
	}, "\n")

	pterm.DefaultBox.
		WithWriter(w).
		WithTitle(" EasyDrop Runtime ").
		WithTitleTopCenter().
		WithHorizontalPadding(2).
		WithVerticalPadding(1).
		WithBoxStyle(pterm.NewStyle(pterm.FgLightCyan)).
		Println(content)
}

func buildValueOrDefault(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}

	return trimmed
}

func ensureDefaultConfigOnStartup(configDir string, logger *log.Logger) error {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return errors.New("config dir is required")
	}
	if logger == nil {
		logger = log.Default()
	}

	configPath := filepath.Join(configDir, "config.yaml")
	info, err := os.Stat(configPath)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("配置文件路径是目录: %s", configPath)
		}
		logger.Printf("检测到配置文件已存在，跳过自动生成: %s", configPath)
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("检查配置文件失败: %w", err)
	}

	if err := config.WriteDefaultConfigFile(configDir); err != nil {
		if errors.Is(err, os.ErrExist) {
			logger.Printf("检测到配置文件已存在，跳过自动生成: %s", configPath)
			return nil
		}
		return fmt.Errorf("自动生成默认配置文件失败: %w", err)
	}

	logger.Printf("配置文件不存在，已自动创建: %s", configPath)
	return nil
}

func ensureJWTKeysOnStartup(configDir string, autoGenerateJWT bool, logger *log.Logger) error {
	if !autoGenerateJWT {
		return nil
	}
	if logger == nil {
		logger = log.Default()
	}

	cfg, err := config.Load(configDir, false)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	if cfg == nil {
		return errors.New("加载配置失败: 配置为空")
	}

	privatePath := strings.TrimSpace(cfg.JWT.PrivateKeyPath)
	publicPath := strings.TrimSpace(cfg.JWT.PublicKeyPath)

	privateExists, err := checkJWTKeyFileExists(privatePath)
	if err != nil {
		return fmt.Errorf("检查 JWT 私钥文件失败: %w", err)
	}
	publicExists, err := checkJWTKeyFileExists(publicPath)
	if err != nil {
		return fmt.Errorf("检查 JWT 公钥文件失败: %w", err)
	}

	switch {
	case privateExists && publicExists:
		logger.Printf("检测到 JWT 密钥文件已存在，跳过自动生成: %s, %s", privatePath, publicPath)
		return nil
	case !privateExists && !publicExists:
		if err := generateJWTTokenPair(privatePath, publicPath, false); err != nil {
			return fmt.Errorf("自动生成 JWT 密钥文件失败: %w", err)
		}
		logger.Printf("JWT 密钥文件不存在，已自动生成: %s, %s", privatePath, publicPath)
		return nil
	default:
		return fmt.Errorf("JWT 密钥文件不完整: %s 与 %s 必须同时存在或同时不存在，请手动修复后重启", privatePath, publicPath)
	}
}

func checkJWTKeyFileExists(path string) (bool, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false, errors.New("路径不能为空")
	}

	info, err := os.Stat(trimmed)
	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("路径是目录: %s", trimmed)
		}
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func generateJWTTokenFiles(outputDir string, forceOverwrite bool) error {
	dir := strings.TrimSpace(outputDir)
	if dir == "" {
		dir = "."
	}

	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")

	return generateJWTTokenPair(privatePath, publicPath, forceOverwrite)
}

func generateJWTTokenPair(privatePath string, publicPath string, forceOverwrite bool) error {
	privatePath = strings.TrimSpace(privatePath)
	if privatePath == "" {
		return errors.New("jwt 私钥路径不能为空")
	}
	publicPath = strings.TrimSpace(publicPath)
	if publicPath == "" {
		return errors.New("jwt 公钥路径不能为空")
	}

	if err := ensureJWTKeyParentDir(privatePath); err != nil {
		return err
	}
	if err := ensureJWTKeyParentDir(publicPath); err != nil {
		return err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("生成 RSA 私钥失败: %w", err)
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("编码 RSA 公钥失败: %w", err)
	}

	if err := writePEMFile(privatePath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey), 0o600, forceOverwrite); err != nil {
		return err
	}
	if err := writePEMFile(publicPath, "PUBLIC KEY", publicDER, 0o644, forceOverwrite); err != nil {
		return err
	}

	return nil
}

func ensureJWTKeyParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 (%s): %w", dir, err)
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
