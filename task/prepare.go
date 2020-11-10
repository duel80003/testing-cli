package task

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"twilio-cli/printer"
)

const (
	path         = "./"
	sendingImage = "sending image..."
)

var (
	client         string
	workflow       string
	language       string = "en"
	country        string
	c              config
	originaldata   originalJSONData
	testingData    TestingData
	defaultContext map[string]string = make(map[string]string)
)
var (
	yellow = printer.ColorInstance("yellow").SprintFunc()
	white  = printer.ColorInstance("white").SprintFunc()
)

type langTranslation map[string]string

type answer map[string][]string

// Config for various condition
type config struct {
	Contexts        []string                   `json:"contexts"`
	ContextsMapping map[string]string          `json:"contextsMapping"`
	RequestInterval int64                      `json:"requestInterval"`
	Translation     map[string]langTranslation `json:"translation"`
}

func (config *config) parseAnswer(o []string) []string {
	result := make([]string, len(o))
	for i, v := range o {
		if defaultContext[v] == "" {
			result[i] = v
		} else {
			result[i] = defaultContext[v]
		}
	}
	return result
}

// OriginalJSONData parse client.json to struct
type originalJSONData struct {
	MediaURL0 string   `json:"mediaURL0"`
	URL       string   `json:"url"`
	From      string   `json:"from"`
	Workflows []answer `json:"answers"`
}

// TestingData is the testable data
type TestingData struct {
	mediaURL0 string
	url       string
	from      string
	answers   []string
}

func (data originalJSONData) getWorkflow() []string {
	for _, v := range data.Workflows {
		if v[workflow] != nil {
			return v[workflow]
		}
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: dialogflow testing tool\n")
	flag.PrintDefaults()
}

func init() {
	flag.StringVar(&client, "c", "", "the name for testing client")
	flag.StringVar(&workflow, "w", "", "the workflow to usage")
	flag.StringVar(&language, "l", "en", "the language to usage")
	flag.StringVar(&country, "ct", "USA", "the country to usage")
}

func checkOptions() {
	if client == "" {
		printer.Error("client is required")
		os.Exit(1)
	}
	if workflow == "" {
		printer.Error("Workflow is required")
		os.Exit(1)
	}
}
func setDefaultContext() {
	defaultContext["image"] = sendingImage
	for _, v := range c.Contexts {
		defaultContext[v] = c.ContextsMapping[fmt.Sprint(v, "_", language)]
	}
	lang := c.Translation[language]
	for k, v := range lang {
		defaultContext[k] = v
	}
}

func readConfig() {
	fileName := client + "_config.json"
	printer.Info("Reading config file: " + yellow(fileName))
	fullPath := path + fileName
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		printer.Error("Reading config file error")
		fmt.Println(err)
	}
	json.Unmarshal(data, &c)
	if c.Translation[strings.ToLower(language)] == nil {
		message := fmt.Sprintf("Translation for %s not provid", strings.ToLower(language))
		printer.Error(message)
		os.Exit(1)
	}
	setDefaultContext()
}

func readOriginalJSONData() {
	fileName := client + ".json"
	printer.Info("Reading test file: " + yellow(fileName))
	fullPath := path + fileName
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		printer.Error("Reading testing data file")
		fmt.Println(err)
	}
	json.Unmarshal(data, &originaldata)
}

func prepareTestingData() {
	w := originaldata.getWorkflow()
	if w == nil {
		message := fmt.Sprintf("workflow %s does not exited", workflow)
		printer.Error(message)
		os.Exit(1)
	}
	testingData.answers = c.parseAnswer(w)
	testingData.from = originaldata.From
	testingData.mediaURL0 = originaldata.MediaURL0
	testingData.url = originaldata.URL
}

// PrepareData for twilio test processing
func PrepareData() {
	flag.Parse()
	checkOptions()
	readConfig()
	readOriginalJSONData()
	prepareTestingData()
	message := "Client: " + yellow(strings.ToUpper(client)) + white(", Test WorkFlow: ") + yellow(workflow)
	printer.Info(message)
	message = "Country: " + yellow(strings.ToUpper(country))
	printer.Info(message)
	fmt.Println("aaa", testingData.answers)
}
