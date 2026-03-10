package cmd

import (
	"context"

	"github.com/spf13/cobra"

	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"

	"github.com/KQAR/feishu-docs/internal/output"
)

func newWikiNodeInfoCmd() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "node",
		Short: "获取节点信息",
		Run: func(cmd *cobra.Command, args []string) {
			req := larkwiki.NewGetNodeSpaceReqBuilder().
				Token(token).
				Build()

			resp, err := larkClient.Wiki.Space.GetNode(context.Background(), req)
			if err != nil {
				output.Errorf("获取节点信息失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取节点信息失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "节点 token (必填)")
	cmd.MarkFlagRequired("token")

	return cmd
}

func newWikiNodeListCmd() *cobra.Command {
	var spaceID, parentNodeToken, pageToken string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "列出知识空间子节点",
		Run: func(cmd *cobra.Command, args []string) {
			builder := larkwiki.NewListSpaceNodeReqBuilder().
				SpaceId(spaceID).
				PageSize(pageSize)

			if parentNodeToken != "" {
				builder.ParentNodeToken(parentNodeToken)
			}
			if pageToken != "" {
				builder.PageToken(pageToken)
			}

			resp, err := larkClient.Wiki.SpaceNode.List(context.Background(), builder.Build())
			if err != nil {
				output.Errorf("获取节点列表失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("获取节点列表失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			if resp.Data == nil || len(resp.Data.Items) == 0 {
				output.Success("暂无节点")
				return
			}

			headers := []string{"节点Token", "类型", "标题", "创建者"}
			var rows [][]string
			for _, item := range resp.Data.Items {
				rows = append(rows, []string{
					deref(item.NodeToken),
					deref(item.ObjType),
					deref(item.Title),
					deref(item.Creator),
				})
			}
			output.Table(headers, rows)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space-id", "s", "", "知识空间 ID (必填)")
	cmd.Flags().StringVarP(&parentNodeToken, "parent", "p", "", "父节点 token (可选，不填则查询根节点)")
	cmd.Flags().IntVarP(&pageSize, "page-size", "n", 20, "每页数量")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "分页标记")
	cmd.MarkFlagRequired("space-id")

	return cmd
}

func newWikiNodeCreateCmd() *cobra.Command {
	var spaceID, objType, parentNodeToken, title, nodeType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建知识空间节点",
		Run: func(cmd *cobra.Command, args []string) {
			nodeBuilder := larkwiki.NewNodeBuilder().
				ObjType(objType).
				NodeType(nodeType)

			if parentNodeToken != "" {
				nodeBuilder.ParentNodeToken(parentNodeToken)
			}
			if title != "" {
				nodeBuilder.Title(title)
			}

			req := larkwiki.NewCreateSpaceNodeReqBuilder().
				SpaceId(spaceID).
				Node(nodeBuilder.Build()).
				Build()

			resp, err := larkClient.Wiki.SpaceNode.Create(context.Background(), req)
			if err != nil {
				output.Errorf("创建节点失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("创建节点失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("节点创建成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space-id", "s", "", "知识空间 ID (必填)")
	cmd.Flags().StringVar(&objType, "obj-type", "docx", "节点对象类型: doc/sheet/mindnote/bitable/file/docx")
	cmd.Flags().StringVarP(&parentNodeToken, "parent", "p", "", "父节点 token")
	cmd.Flags().StringVarP(&title, "title", "t", "", "节点标题")
	cmd.Flags().StringVar(&nodeType, "node-type", "origin", "节点类型: origin/shortcut")
	cmd.MarkFlagRequired("space-id")

	return cmd
}

func newWikiNodeMoveCmd() *cobra.Command {
	var spaceID, nodeToken, targetParentToken, targetSpaceID string

	cmd := &cobra.Command{
		Use:   "move",
		Short: "移动知识空间节点",
		Long:  "移动节点到指定位置，支持跨知识空间移动。子节点会一起移动。",
		Run: func(cmd *cobra.Command, args []string) {
			bodyBuilder := larkwiki.NewMoveSpaceNodeReqBodyBuilder().
				TargetParentToken(targetParentToken)

			if targetSpaceID != "" {
				bodyBuilder.TargetSpaceId(targetSpaceID)
			}

			req := larkwiki.NewMoveSpaceNodeReqBuilder().
				SpaceId(spaceID).
				NodeToken(nodeToken).
				Body(bodyBuilder.Build()).
				Build()

			resp, err := larkClient.Wiki.SpaceNode.Move(context.Background(), req)
			if err != nil {
				output.Errorf("移动节点失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("移动节点失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("节点移动成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space-id", "s", "", "当前知识空间 ID (必填)")
	cmd.Flags().StringVarP(&nodeToken, "node-token", "t", "", "要移动的节点 token (必填)")
	cmd.Flags().StringVar(&targetParentToken, "target-parent", "", "目标父节点 token (必填)")
	cmd.Flags().StringVar(&targetSpaceID, "target-space", "", "目标知识空间 ID (跨空间移动时填写)")
	cmd.MarkFlagRequired("space-id")
	cmd.MarkFlagRequired("node-token")
	cmd.MarkFlagRequired("target-parent")

	return cmd
}

func newWikiNodeCopyCmd() *cobra.Command {
	var spaceID, nodeToken, targetParentToken, targetSpaceID, title string

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "复制知识空间节点",
		Run: func(cmd *cobra.Command, args []string) {
			bodyBuilder := larkwiki.NewCopySpaceNodeReqBodyBuilder().
				TargetParentToken(targetParentToken).
				TargetSpaceId(targetSpaceID)

			if title != "" {
				bodyBuilder.Title(title)
			}

			req := larkwiki.NewCopySpaceNodeReqBuilder().
				SpaceId(spaceID).
				NodeToken(nodeToken).
				Body(bodyBuilder.Build()).
				Build()

			resp, err := larkClient.Wiki.SpaceNode.Copy(context.Background(), req)
			if err != nil {
				output.Errorf("复制节点失败: %v", err)
			}
			if !resp.Success() {
				output.Errorf("复制节点失败 [%d]: %s (request_id: %s)", resp.Code, resp.Msg, resp.RequestId())
			}

			output.Success("节点复制成功")
			output.JSON(resp.Data)
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space-id", "s", "", "源知识空间 ID (必填)")
	cmd.Flags().StringVarP(&nodeToken, "node-token", "t", "", "要复制的节点 token (必填)")
	cmd.Flags().StringVar(&targetParentToken, "target-parent", "", "目标父节点 token (必填)")
	cmd.Flags().StringVar(&targetSpaceID, "target-space", "", "目标知识空间 ID (必填)")
	cmd.Flags().StringVar(&title, "title", "", "复制后的新标题")
	cmd.MarkFlagRequired("space-id")
	cmd.MarkFlagRequired("node-token")
	cmd.MarkFlagRequired("target-parent")
	cmd.MarkFlagRequired("target-space")

	return cmd
}
