package cmd

import (
	"context"

	"github.com/spf13/cobra"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"

	"github.com/KQAR/feishu-docs/internal/output"
)

func newWikiSpaceListCmd() *cobra.Command {
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "列出知识空间",
		Run: func(cmd *cobra.Command, args []string) {
			builder := larkwiki.NewListSpaceReqBuilder().
				PageSize(pageSize)
			if pageToken != "" {
				builder.PageToken(pageToken)
			}

			resp, err := larkClient.Wiki.Space.List(context.Background(), builder.Build())
			if err != nil {
				output.Errorf("获取知识空间列表失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取知识空间列表失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			if resp.Data == nil || len(resp.Data.Items) == 0 {
				output.Success("暂无知识空间")
				return
			}

			headers := []string{"空间ID", "名称", "描述", "可见性"}
			var rows [][]string
			for _, item := range resp.Data.Items {
				rows = append(rows, []string{
					deref(item.SpaceId),
					deref(item.Name),
					deref(item.Description),
					deref(item.Visibility),
				})
			}
			output.Table(headers, rows)
		},
	}

	cmd.Flags().IntVarP(&pageSize, "page-size", "n", 20, "每页数量")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "分页标记")

	return cmd
}

func newWikiSpaceGetCmd() *cobra.Command {
	var spaceID string

	cmd := &cobra.Command{
		Use:   "space",
		Short: "获取知识空间信息",
		Run: func(cmd *cobra.Command, args []string) {
			req := larkwiki.NewGetSpaceReqBuilder().
				SpaceId(spaceID).
				Build()

			resp, err := larkClient.Wiki.Space.Get(context.Background(), req)
			if err != nil {
				output.Errorf("获取知识空间信息失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取知识空间信息失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "id", "i", "", "知识空间 ID (必填)")
	cmd.MarkFlagRequired("id")

	return cmd
}
