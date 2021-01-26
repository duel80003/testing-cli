package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"twilio-test-cli/logger"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type dialogflowClient struct {
	ProjectID   string
	SessionID   string
	Credentials []byte
	User        string
}

func (client *dialogflowClient) init() {
	logger.Info("Init dialogflow client..")
	configData := readJSONFile(createFilePath([]string{dialogflowFlag.Client, "_config.json"}))
	dialogflowProfile := gjson.GetBytes(configData, ConcatString([]string{"dialogflow.", dialogflowFlag.Environment})).Map()
	client.ProjectID = dialogflowProfile["projectId"].Str
	client.User = dialogflowProfile["user"].Str
	client.SessionID = getSessionID()
	credentialMap := map[string]string{
		"type":         dialogflowProfile["type"].Str,
		"client_email": dialogflowProfile["client_email"].Str,
		"private_key":  dialogflowProfile["private_key"].Str,
	}
	bytesData, err := json.Marshal(credentialMap)
	if err != nil {
		logger.Error("create credential failure")
		os.Exit(1)
	}
	client.Credentials = bytesData
}

func (client *dialogflowClient) sendToDialogflow() (string, error) {
	ctx := context.Background()
	sessionClient, err := dialogflow.NewSessionsClient(ctx, option.WithCredentialsJSON(client.Credentials))

	if err != nil {
		return "", err
	}
	defer sessionClient.Close()
	sessionPath := client.getSessionPath()
	// logger.Info("sessionPath: " + sessionPath)
	textInput := dialogflowpb.TextInput{Text: dialogflowFlag.Text, LanguageCode: dialogflowFlag.Language}
	queryTextInput := dialogflowpb.QueryInput_Text{Text: &textInput}
	queryInput := dialogflowpb.QueryInput{Input: &queryTextInput}
	request := dialogflowpb.DetectIntentRequest{
		Session:    sessionPath,
		QueryInput: &queryInput,
	}
	if dialogflowFlag.Context != "" {
		context := &dialogflowpb.Context{}
		context.LifespanCount = 3
		context.Name = ConcatString([]string{sessionPath, "/contexts/", dialogflowFlag.Context})
		logger.Info(context.Name)
		contexts := []*dialogflowpb.Context{context}
		queryParameters := &dialogflowpb.QueryParameters{Contexts: contexts}
		request.QueryParams = queryParameters
	}
	response, err := sessionClient.DetectIntent(ctx, &request)
	if err != nil {
		return "", err
	}
	queryResult := response.GetQueryResult()
	fulfillmentText := queryResult.GetFulfillmentText()
	return fulfillmentText, nil
}

func (client *dialogflowClient) deleteContext() error {
	ctx := context.Background()
	contextClient, err := dialogflow.NewContextsClient(ctx, option.WithCredentialsJSON(client.Credentials))
	if err != nil {
		return err
	}
	req := &dialogflowpb.DeleteAllContextsRequest{
		Parent: client.getSessionPath(),
	}
	deleteContextError := contextClient.DeleteAllContexts(ctx, req)
	if deleteContextError != nil {
		return err
	}
	return nil
}

func (client *dialogflowClient) getSessionPath() string {
	switch dialogflowFlag.Environment {
	case "dev":
		return fmt.Sprintf("projects/%s/agent/sessions/%s", client.ProjectID, client.SessionID)
	default:
		return fmt.Sprintf("projects/%s/agent/environments/%s/users/%s/sessions/%s", client.ProjectID, dialogflowFlag.Environment, client.User, client.SessionID)
	}
}

var client dialogflowClient

func getSessionID() string {
	clientData := readJSONFile(createFilePath([]string{dialogflowFlag.Client, ".json"}))
	return gjson.GetBytes(clientData, "from").Str
}

var cmdDialogFlow = &cobra.Command{
	Use:   "dialogflow",
	Short: "delete Context of Dialogflow",
	Long:  `delete the current user context, it can simulate context expired`,
}

var cmdText = &cobra.Command{
	Use:   "text",
	Short: "send a text to Dialogflow",
	Long:  `Directly send a message to Dialogflow agent for test its result`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info(ConcatString([]string{"Client: ", dialogflowFlag.Client}))
		logger.Info(ConcatString([]string{"Language: ", dialogflowFlag.Language}))
		logger.Info(ConcatString([]string{"Message: ", dialogflowFlag.Text}))

		client.init()
		result, err := client.sendToDialogflow()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		logger.Info(result)
	},
}

var cmdDeleteAllContexts = &cobra.Command{
	Use:   "deleteAllContexts",
	Short: "delete context",
	Long:  `delete dialogflow context by given session id`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info(ConcatString([]string{"Client: ", dialogflowFlag.Client}))
		logger.Info(ConcatString([]string{"Language: ", dialogflowFlag.Language}))

		client.init()
		err := client.deleteContext()
		if err != nil {
			logger.Error("Delete dialogflwo contexts failure")
		}
		logger.Info("Delete contexts success")
	},
}
