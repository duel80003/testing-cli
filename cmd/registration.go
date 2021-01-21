package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"twilio-test-cli/logger"

	"github.com/levigross/grequests"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

const path = "./"
const image string = "image"
const text string = "text"

var yellow = logger.ColorInstance("yellow").SprintFunc()

//TestData contain necessary properties for test registration flow
type TestData struct {
	Image           string
	From            string
	URL             string
	answers         []string
	RequestInterval int
}

func (t *TestData) smsRequestBody(answer string) map[string]string {
	return map[string]string{
		"AccountSid": "test",
		"From":       t.From,
		"To":         "test",
		"Body":       answer,
		"FromState":  "FL",
	}
}

func (t *TestData) makeSMSRequest(answer string) (*grequests.Response, error) {
	logger.ShowAnswer(answer)
	requestBody := &grequests.RequestOptions{
		JSON: t.smsRequestBody(answer),
	}
	return grequests.Post(testData.URL, requestBody)
}

func (t *TestData) makeMMSRequest() (*grequests.Response, error) {
	logger.ShowAnswer("Sending image...")
	requestBody := &grequests.RequestOptions{
		JSON: t.mmsRequestBody(),
	}
	return grequests.Post(testData.URL, requestBody)
}

func (t *TestData) mmsRequestBody() map[string]string {
	return map[string]string{
		"AccountSid":        "test",
		"From":              t.From,
		"To":                "test",
		"Body":              "",
		"FromState":         "FL",
		"NumMedia":          "1",
		"MediaUrl0":         t.Image,
		"MediaContentType0": "jpg",
	}
}

type xmlResponse struct {
	Message string `xml:"Message"`
}

type xmlMediaResponse struct {
	Message mediaXML `xml:"Message"`
}

type mediaXML struct {
	Body  string   `xml:"Body"`
	Media []string `xml:"Media"`
}

var testData TestData

var cmdRegistration = &cobra.Command{
	Use:   "regis",
	Short: "Test SMS registration flow",
	Long:  `Read the json file by client flag, and test the complete registration process`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Registration start")
		logger.Info(ConcatString([]string{"Client: ", twilioFlag.Client}))
		logger.Info(ConcatString([]string{"Workflow: ", twilioFlag.Workflow}))
		logger.Info(ConcatString([]string{"Language: ", twilioFlag.Language}))
		logger.Info(ConcatString([]string{"Country: ", twilioFlag.Country}))
		prepareTestData(&twilioFlag)
		startTestProcess()
		logger.Info("Test End")

	},
}

func createFilePath(args []string) string {
	fileName := ConcatString(args)
	logger.Info("Reading config file: " + yellow(fileName))
	return ConcatString([]string{path, fileName})
}

func findAnswersByWorkflow(flag *Flag, objectArray []gjson.Result) []gjson.Result {
	workflow := flag.Workflow
	for _, v := range objectArray {
		m := v.Map()
		if m[workflow].Type != gjson.Null {
			return m[workflow].Array()
		}
	}
	return nil
}

func readJSONFile(fullPath string) []byte {
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		logger.Error(ConcatString([]string{"Read file", fullPath, " failure."}))
		os.Exit(1)
	}
	return data
}

func prepareTestData(flag *Flag) {
	logger.Info("Perparing test data...")
	clientData := readJSONFile(createFilePath([]string{flag.Client, ".json"}))
	configData := readJSONFile(createFilePath([]string{flag.Client, "_config.json"}))

	testData.RequestInterval = int(gjson.GetBytes(configData, "requestInterval").Num)
	queryString := gjson.GetBytes(configData, ConcatString([]string{"queryString.", strings.ToLower(flag.Country)}))
	testData.Image = gjson.GetBytes(clientData, ConcatString([]string{"media_", flag.Image})).Str
	testData.From = gjson.GetBytes(clientData, "from").Str
	url := gjson.GetBytes(clientData, "url").Str
	var questionMark string
	if queryString.Type.String() == "String" && strings.HasPrefix("?", queryString.Str) {
		questionMark = ""
	} else {
		questionMark = "?"
	}
	testData.URL = ConcatString([]string{url, questionMark, queryString.Str})
	answers := findAnswersByWorkflow(flag, gjson.GetBytes(clientData, "answers").Array())
	if answers == nil {
		logger.Error("Inexistent workflow")
		os.Exit(1)
	}
	testData.answers = composeAnswersAndContext(flag, answers, configData)
}

func composeAnswersAndContext(flag *Flag, answers []gjson.Result, configData []byte) []string {
	contexts := gjson.GetBytes(configData, "contexts").Array()
	contextsMapping := gjson.GetBytes(configData, "contextsMapping").Map()
	translation := gjson.GetBytes(configData, ConcatString([]string{"translation.", flag.Language})).Map()

	result := make([]string, len(answers))
	contextMap := make(map[string]string)
	for _, v := range contexts {
		contextMap[v.Str] = contextsMapping[ConcatString([]string{v.Str, "_", flag.Language})].Str
	}
	for i, v := range answers {
		answer := v.Str
		if contextMap[answer] != "" {
			result[i] = contextMap[answer]
		} else if translation[answer].Str != "" {
			result[i] = translation[answer].Str
		} else {
			result[i] = answer
		}
	}
	return result
}

func startTestProcess() {
	logger.Info("Start testing...")
	logger.Divider()
	var printHTTPError = func(err error) {
		logger.Error("Http post error")
		logger.Print(err)
	}
	for _, v := range testData.answers {
		var messageType string
		if v == "image" {
			messageType = image
			response, err := testData.makeMMSRequest()
			if err != nil {
				printHTTPError(err)
			}
			displayResult(messageType, response)
		} else {
			messageType = text
			response, err := testData.makeSMSRequest(v)
			if err != nil {
				printHTTPError(err)
			}
			displayResult(messageType, response)
		}
		time.Sleep(time.Duration(testData.RequestInterval) * time.Millisecond)
	}
}

func displayResult(messageType string, response *grequests.Response) {
	if messageType == text {
		xmlResponse := &xmlResponse{}
		response.XML(xmlResponse, nil)
		logger.ShowQuestion(xmlResponse.Message)
	} else {
		xmlResponse := &xmlMediaResponse{}
		response.XML(xmlResponse, nil)
		if xmlResponse.Message.Body == "" {
			logger.ShowQuestion("Empty message")
		} else {
			logger.ShowQuestion(xmlResponse.Message.Body)
		}
		for _, url := range xmlResponse.Message.Media {
			yellow := logger.ColorInstance("yellow").SprintFunc()
			magenta := logger.ColorInstance("magenta")
			magenta.Printf("%-18s%s \n", "Media URL:", yellow(url))
		}
	}
	logger.Divider()
}
