package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultConfigPath = "config/feishu-docs/config.json"

// Config 飞书应用配置
type Config struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// Load 从 ~/config/feishu-docs/config.json 加载配置
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	configPath := filepath.Join(homeDir, defaultConfigPath)
	return LoadFrom(configPath)
}

// LoadFrom 从指定路径加载配置文件
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// EnsureConfigFile 确保配置文件存在，不存在时创建模板
func EnsureConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}

	configPath := filepath.Join(homeDir, defaultConfigPath)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %w", err)
	}

	template := Config{
		AppID:     "your_app_id_here",
		AppSecret: "your_app_secret_here",
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置模板失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return "", fmt.Errorf("写入配置模板失败: %w", err)
	}

	return configPath, nil
}

func (c *Config) validate() error {
	if c.AppID == "" || c.AppID == "your_app_id_here" {
		return fmt.Errorf("请在配置文件中设置有效的 app_id")
	}
	if c.AppSecret == "" || c.AppSecret == "your_app_secret_here" {
		return fmt.Errorf("请在配置文件中设置有效的 app_secret")
	}
	return nil
}
