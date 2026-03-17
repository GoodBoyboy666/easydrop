package storage

import "time"

const (
	BackendLocal = "local"
	BackendS3    = "s3"
)

// Config 定义存储总配置。
type Config struct {
	Backend string      `mapstructure:"backend" yaml:"backend"`
	Local   LocalConfig `mapstructure:"local" yaml:"local"`
	S3      S3Config    `mapstructure:"s3" yaml:"s3"`
}

// LocalConfig 定义本地文件存储配置。
type LocalConfig struct {
	BasePath string `mapstructure:"base_path" yaml:"base_path"`
	BaseURL  string `mapstructure:"base_url" yaml:"base_url"`
}

// S3Config 定义 S3 存储配置。
type S3Config struct {
	Bucket        string        `mapstructure:"bucket" yaml:"bucket"`
	Region        string        `mapstructure:"region" yaml:"region"`
	AccessKey     string        `mapstructure:"access_key" yaml:"access_key"`
	SecretKey     string        `mapstructure:"secret_key" yaml:"secret_key"`
	Endpoint      string        `mapstructure:"endpoint" yaml:"endpoint"`
	UsePathStyle  bool          `mapstructure:"use_path_style" yaml:"use_path_style"`
	PublicBaseURL string        `mapstructure:"public_base_url" yaml:"public_base_url"`
	PresignExpire time.Duration `mapstructure:"presign_expire" yaml:"presign_expire"`
}
