package cmd

import "github.com/spf13/cobra"

func newWikiCmd() *cobra.Command {
	wikiCmd := &cobra.Command{
		Use:   "wiki",
		Short: "知识库管理",
		Long:  "管理飞书知识库(Wiki)，支持知识空间和节点的增删改查操作。",
	}

	wikiCmd.AddCommand(
		newWikiSpaceListCmd(),
		newWikiSpaceGetCmd(),
		newWikiNodeInfoCmd(),
		newWikiNodeListCmd(),
		newWikiNodeCreateCmd(),
		newWikiNodeMoveCmd(),
		newWikiNodeCopyCmd(),
	)

	return wikiCmd
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
