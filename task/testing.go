package task

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
	"twilio-cli/printer"
)

type xmlResponse struct {
	Message string `xml:"Message"`
}

type xmlMediaResponse struct {
	mediaXML `xml:"Message"`
}

type mediaXML struct {
	Body  string   `xml:"Body"`
	Media []string `xml:"Media"`
}

// Testing is request body
type Testing struct {
	TestingData
}

//NewInstance create a Testing instance
func NewInstance() *Testing {
	return &Testing{testingData}
}

// Start the testing program.
func (t *Testing) Start() {
	printer.Info("Test Start")
	printer.Divider()
	t.startTestProcess()
	printer.Info("Test End")
}

//RequestBody sms request body
func (t *Testing) requestBody(question string) []byte {
	body, _ := json.Marshal(map[string]string{
		"AccountSid": "test",
		"From":       t.from,
		"To":         "test",
		"Body":       question,
		"FromState":  "FL",
	})
	return body
}

// MMSRequestBody mms request body
func (t *Testing) mmsRequestBody() []byte {
	body, _ := json.Marshal(map[string]string{
		"AccountSid":        "test",
		"From":              t.from,
		"To":                "test",
		"Body":              "",
		"FromState":         "FL",
		"NumMedis":          "1",
		"MediaUrl0":         t.mediaURL0,
		"MediaContentType0": "jpg",
	})
	return body
}

func (t *Testing) startTestProcess() {
	var body []byte
	for _, value := range t.answers {
		if value != sendingImage {
			printer.ShowAnswer(value)
			body = t.requestBody(value)
		} else {
			printer.ShowAnswer("Sending image...")
			cyan := printer.ColorInstance("cyan").SprintFunc()
			magenta := printer.ColorInstance("magenta")
			magenta.Printf("%-18s%s \n", "Image URL:", cyan(t.mediaURL0))
			body = t.mmsRequestBody()
		}
		res, err := http.Post(t.url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			printer.Error("Http post error")
			printer.Print(err)
		}
		body, _ = ioutil.ReadAll(res.Body)
		t.processingResponse(body)
	}
}

func (t *Testing) processingResponse(body []byte) {
	responseXML := &xmlResponse{}
	_ = xml.Unmarshal(body, &responseXML)
	var basicDelayTime time.Duration = time.Duration(c.RequestInterval)
	if responseXML.Message != "" {
		printer.ShowQuestion(responseXML.Message)
	} else {
		responseXML := &xmlMediaResponse{}
		_ = xml.Unmarshal(body, &responseXML)
		if responseXML.Body == "" {
			printer.ShowQuestion("Dialogflow timeout, empty message")
		} else {
			printer.ShowQuestion(responseXML.Body)
		}
		for _, url := range responseXML.Media {
			yellow := printer.ColorInstance("yellow").SprintFunc()
			magenta := printer.ColorInstance("magenta")
			magenta.Printf("%-18s%s \n", "Media URL:", yellow(url))
		}
	}
	printer.Divider()
	time.Sleep(basicDelayTime * time.Millisecond)
}
