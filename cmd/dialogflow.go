package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdDialgoFlow = &cobra.Command{
	Use:   "deleteContext",
	Short: "delete Context of Dialogflow",
	Long:  `delete the current user context, it can simulate context expired`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aaaa")
	},
}
