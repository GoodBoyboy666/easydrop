package avatar

const (
	// DefaultGravatarBaseURL 是 Gravatar 默认官方地址。
	DefaultGravatarBaseURL = "https://www.gravatar.com/avatar/"
)

// Config 是头像相关配置。
type Config struct {
	GravatarBaseURL string `mapstructure:"gravatar_base_url" yaml:"gravatar_base_url"`
}
