package version

import "fmt"

// 通过 ldflags 在编译时注入
var (
	Version   = "0.3.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Full 返回完整版本信息
func Full() string {
	return fmt.Sprintf("feishu-docs-cli %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
