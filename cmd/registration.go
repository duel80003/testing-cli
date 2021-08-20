package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"twilio-test-cli/logger"

	"github.com/levigross/grequests"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

const image string = "image"

var yellow = logger.ColorInstance("yellow").SprintFunc()

//TestData contain necessary properties for test registration flow
type TestData struct {
	Image           string
	From            string
	URL             string
	answers         []string
	RequestInterval int
	BaseURL         string
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
		start := make(chan bool)
		delete := make(chan bool)
		done := make(chan bool)
		go prepareTestData(&twilioFlag, delete)
		go deleteSession(delete, start, done)
		resChan := make(chan *grequests.Response)
		go startTestProcess(resChan, done, start)
		go displayResult(resChan)
		<-done
		logger.Info("Test End")
	},
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

func prepareTestData(flag *Flag, delete chan<- bool) {
	logger.Info("Preparing test data...")
	clientData := readJSONFile(createFilePath([]string{flag.Client, ".json"}))
	client := strings.Split(flag.Client, "_")
	configData := readJSONFile(createFilePath([]string{client[0], "_config.json"}))

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
	testData.BaseURL = url
	testData.URL = ConcatString([]string{url, questionMark, queryString.Str})
	answers := findAnswersByWorkflow(flag, gjson.GetBytes(clientData, "answers").Array())
	if answers == nil {
		logger.Error(fmt.Sprintf("workflow %s not exist", flag.Workflow))
		os.Exit(1)
	}
	testData.answers = translationContextMapping(flag, answers, configData)
	delete <- true
	close(delete)
}

func deleteSession(delete <-chan bool, start chan<- bool, done chan<- bool) {
	for val := range delete {
		if val {
			json, _ := json.Marshal(map[string]string{
				"userId": testData.From,
			})
			url := testData.BaseURL + "/delete-session"
			logger.Info("Delete session url: " + url)
			requestBody := &grequests.RequestOptions{
				JSON: json,
			}
			res, err := grequests.Delete(url, requestBody)
			if err != nil {
				logger.Error(url + "is unavailable")
			}
			type deleteRes struct {
				Message string `json:"message"`
			}
			d := &deleteRes{}
			res.JSON(d)
			s := res.StatusCode
			if s == 200 {
				logger.Info("Delete session response: " + d.Message)
				start <- true
				close(start)
			} else {
				logger.Info("Delete session err: " + d.Message)
				done <- false
				close(start)
				close(done)
			}
		} else {
			break
		}
	}
}

func translationContextMapping(flag *Flag, answers []gjson.Result, configData []byte) []string {
	translation := gjson.GetBytes(configData, ConcatString([]string{"translation_context.", flag.Language})).Map()
	result := make([]string, len(answers))
	contextMap := make(map[string]string)
	for i, v := range answers {
		answer := strings.ToLower(v.Str)
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

func startTestProcess(resChan chan<- *grequests.Response, done chan<- bool, start <-chan bool) {
	for s := range start {
		if s {
			logger.Info("Start testing...")
			logger.Divider()
			var printHTTPError = func(err error) {
				logger.Error("Http post error")
				logger.Print(err)
			}
			for _, v := range testData.answers {
				if v == image {
					response, err := testData.makeMMSRequest()
					if err != nil {
						printHTTPError(err)
					}
					resChan <- response
				} else {
					response, err := testData.makeSMSRequest(v)
					if err != nil {
						printHTTPError(err)
					}
					resChan <- response
				}
				time.Sleep(time.Duration(testData.RequestInterval) * time.Millisecond)
			}
			done <- true
			close(resChan)
			close(done)
		} else {
			done <- false
			close(done)
			break
		}
	}
}

func displayResult(resChan <-chan *grequests.Response) {
	for response := range resChan {
		xmlResponse := &xmlResponse{}
		bytes := response.Bytes()
		response.XML(xmlResponse, nil)

		if xmlResponse.Message != "" {
			logger.ShowQuestion(xmlResponse.Message)
		} else {
			xmlMediaResponse := &xmlMediaResponse{}
			_ = xml.Unmarshal(bytes, &xmlMediaResponse)
			response.ClearInternalBuffer()
			if xmlMediaResponse.Message.Body == "" {
				logger.ShowQuestion("Empty message")
				os.Exit(1)
			} else {
				logger.ShowQuestion(xmlMediaResponse.Message.Body)
			}
			for _, url := range xmlMediaResponse.Message.Media {
				yellow := logger.ColorInstance("yellow").SprintFunc()
				magenta := logger.ColorInstance("magenta")
				magenta.Printf("%-18s%s \n", "Media URL:", yellow(url))
			}
		}
		response.Close()
		logger.Divider()
	}
}
