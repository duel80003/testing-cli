package cmd

import (
	"github.com/spf13/cobra"
)

// Flag contain twilio flag for test SMS process
type Flag struct {
	Client   string
	Workflow string
	Language string
	Country  string
	Image    string
}

//DialogflowFlag for dialogflow test options
type DialogflowFlag struct {
	Text        string
	Client      string
	Language    string
	Environment string
	Context     string
}

var twilioFlag Flag
var dialogflowFlag DialogflowFlag

var rootCmd = &cobra.Command{
	Use:   "twilio",
	Short: "Twilio is a CLI for test sms registration flow",
	Long: `Setup test data by json, test the program is working,
it's for developer without testable phone number.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// Do Stuff Here
	// fmt.Print("Start test process \n")
	// },
	Version: "v1.0.2",
}

func init() {
	cmdRegistration.PersistentFlags().StringVarP(&twilioFlag.Client, "client", "c", "", "要測試的客戶")
	cmdRegistration.PersistentFlags().StringVarP(&twilioFlag.Workflow, "workflow", "w", "", "要測試的流程")
	cmdRegistration.PersistentFlags().StringVarP(&twilioFlag.Language, "language", "l", "en", "要測試的語言")
	cmdRegistration.PersistentFlags().StringVarP(&twilioFlag.Country, "country", "C", "US", "要測試的國家")
	cmdRegistration.PersistentFlags().StringVarP(&twilioFlag.Image, "image", "i", "0", "要測試的圖片")

	cmdRegistration.MarkPersistentFlagRequired("client")
	cmdRegistration.MarkPersistentFlagRequired("workflow")

	rootCmd.AddCommand(cmdRegistration)
}

// Execute command line
func Execute() {

	rootCmd.Execute()
}
