//go:build !dev
// +build !dev

package config

// getDefaultConfig 生产环境配置（不使用 dev build tag 时）
func getDefaultConfig() *Config {
	return getProdConfig()
}
