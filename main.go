package main

import (
	"context"
	"easydrop/internal/di"
	"easydrop/internal/router"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	var configDir string
	flag.StringVar(&configDir, "config-dir", "", "config directory containing config.yaml")
	flag.Parse()

	strict := false
	if configDir == "" {
		configDir = "data"
	} else {
		strict = true
	}

	app, err := di.Initialize(configDir, strict)
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
