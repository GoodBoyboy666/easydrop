package main

import (
	"easydrop/internal/di"
	"flag"
	"log"
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
	_ = app

	// TODO: 初始化并启动应用。
}
