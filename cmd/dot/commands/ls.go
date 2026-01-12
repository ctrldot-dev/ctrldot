package commands

import (
	"fmt"
	"time"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/spf13/cobra"
)

func newLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls <node-id>",
		Short: "List children of a node",
		Long:  "List child nodes connected via PARENT_OF links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := getContext(cmd)
			if err != nil {
				return err
			}

			nodeID := args[0]
			asofSeq, _ := cmd.Flags().GetInt64("asof-seq")
			asofTimeStr, _ := cmd.Flags().GetString("asof-time")

			// Expand with depth=1
			req := client.ExpandRequest{
				IDs:         []string{nodeID},
				NamespaceID: getNamespaceID(ctx.Config),
				Depth:       1,
				AsOfSeq:     asofSeq,
			}

			if asofTimeStr != "" {
				t, err := time.Parse(time.RFC3339, asofTimeStr)
				if err != nil {
					return err
				}
				req.AsOfTime = &t
			}

			result, err := ctx.Client.Expand(req)
			if err != nil {
				handleError(err)
				return nil
			}

			// Filter children (PARENT_OF links where from == nodeID)
			var children []string
			for _, link := range result.Links {
				if link.Type == "PARENT_OF" && link.FromNodeID == nodeID {
					children = append(children, link.ToNodeID)
				}
			}

			// Print children
			json, _ := cmd.Flags().GetBool("json")
			if json {
				fmt.Printf(`{"children": %q}\n`, children)
			} else {
				for _, childID := range children {
					// Find child node to get title
					var title string
					for _, node := range result.Nodes {
						if node.ID == childID {
							title = node.Title
							break
						}
					}
					if title != "" {
						fmt.Printf("%s  %s\n", childID, title)
					} else {
						fmt.Println(childID)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64("asof-seq", 0, "As-of sequence number")
	cmd.Flags().String("asof-time", "", "As-of time (ISO format)")

	return cmd
}
