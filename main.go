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

	cmd := &cobra.Command{
		Use:           "easydrop",
		Short:         "EasyDrop 服务端程序",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: strings.Join([]string{
			"easydrop",
			"easydrop --config-dir data",
			"easydrop generate-jwt-token data/jwt --force",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(configDir)
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.Flags().StringVar(&configDir, "config-dir", "", "config directory containing config.yaml")
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

func runServer(configDir string) error {
	if configDir == "" {
		configDir = "data"
	}

	printBuildInfoBanner(os.Stdout)

	app, err := di.Initialize(configDir, false)
	if err != nil {
		return fmt.Errorf("初始化应用失败: %w", err)
	}
	if app == nil || app.Config == nil {
		return errors.New("初始化应用失败: 应用配置为空")
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
