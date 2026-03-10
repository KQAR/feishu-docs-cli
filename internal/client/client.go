package client

import (
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/KQAR/feishu-docs/internal/config"
)

// New 根据配置创建飞书 SDK 客户端
func New(cfg *config.Config) *lark.Client {
	return lark.NewClient(cfg.AppID, cfg.AppSecret)
}
