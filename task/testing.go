package task

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"twilio-cli/printer"
)

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

// Testing is request body
type Testing struct {
	TestingData
}

//NewInstance create a Testing instance
func NewInstance() *Testing {
	return &Testing{testingData}
}

func (t *Testing) deleteSession() {
	requestBody, _ := json.Marshal(map[string]string{
		"userId": t.from,
	})

	url := t.originalURL + "/delete-session"
	printer.Info("Delete session url: " + url)
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		printer.Error("Delete session failure")
		os.Exit(1)
	}
	s := res.StatusCode
	body, _ := ioutil.ReadAll(res.Body)
	var result map[string]string
	_ = json.Unmarshal(body, &result)
	if s == 200 {
		printer.Info("Delete session response: " + result["message"])
	} else {
		printer.Info("Delete session err: " + result["message"])
	}
	time.Sleep(3000 * time.Millisecond)
}

// Start the testing program.
func (t *Testing) Start() {
	t.deleteSession()
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
		"NumMedia":          "1",
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
		if responseXML.Message.Body == "" {
			printer.ShowQuestion("Empty message")
		} else {
			printer.ShowQuestion(responseXML.Message.Body)
		}
		for _, url := range responseXML.Message.Media {
			yellow := printer.ColorInstance("yellow").SprintFunc()
			magenta := printer.ColorInstance("magenta")
			magenta.Printf("%-18s%s \n", "Media URL:", yellow(url))
		}
	}
	printer.Divider()
	time.Sleep(basicDelayTime * time.Millisecond)
}
