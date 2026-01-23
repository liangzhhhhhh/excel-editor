//go:build dev
// +build dev

package config

// getDefaultConfig 开发环境配置（使用 dev build tag 时）
func getDefaultConfig() *Config {
	return getDevConfig()
}
