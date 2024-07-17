package generate

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Seann-Moser/wp/source_code"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

var OllamaModelllama3 = "llama3"
var OllamaModelDeepSeekCoderV2 = "deepseek-coder-v2"

var AllOllamaModels = []string{OllamaModelllama3, OllamaModelDeepSeekCoderV2}

// Grab all of the data in:
// - "Tags:"
// - "Language:"
// - "Artists:"
//   - links to go to that webpage will be associated to the link and will only be a partial path
var prompt = `
## Source Code
'''
%s
'''

## Rules
- Images will end with (".png",".jpg",".webp")
- Images will not always be in the <img> tag
- Images can be in any tag or child tag of a parent tag with "image"" any where in the id or name 
- Links can be in the href elements ("src","href","data")

## Question
Can you provide me a list to all images in the website?
`
var _ Generator = &OllamaClient{}

type OllamaClient struct {
	client               *http.Client
	hostURL              string
	sourceCode           source_code.SourceGetter
	model                string
	externalFunctions    []*ExternalFunctions
	externalFunctionsMap map[string]*ExternalFunctions
}

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type Response struct {
	Model     string    `json:"model"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

func OllamaFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("ollama", pflag.ExitOnError)
	fs.String("ollama-host-url", "http://localhost:8081", "Host URL")
	return fs
}

func NewOllamaFlags(client *http.Client, sourceCode source_code.SourceGetter) *OllamaClient {
	return &OllamaClient{
		client:               client,
		hostURL:              viper.GetString("ollama-host-url"),
		sourceCode:           sourceCode,
		model:                "llama2-uncensored",
		externalFunctionsMap: make(map[string]*ExternalFunctions),
	}
}

func NewOllama(client *http.Client, hostURL, model string, sourceCode source_code.SourceGetter) *OllamaClient {
	return &OllamaClient{
		client:               client,
		hostURL:              hostURL,
		sourceCode:           sourceCode,
		model:                model,
		externalFunctions:    nil,
		externalFunctionsMap: make(map[string]*ExternalFunctions),
	}
}

func (o *OllamaClient) Ping(ctx context.Context) error {
	return nil
}

func (o *OllamaClient) GenerateParser(ctx context.Context, url string) error {
	data, _, err := o.sourceCode.Get(ctx, url)
	if err != nil {
		return err
	}
	//println(string(data))
	u := o.hostURL + "/api/generate"
	r := Request{
		Model:  o.model,
		Prompt: fmt.Sprintf(prompt, string(data)),
	}
	requestBody, err := json.Marshal(r)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	text := ""
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		r := Response{}
		err = json.Unmarshal([]byte(scanner.Text()), &r)
		if err != nil {
			return err
		}
		text += r.Response
	}
	println(text)
	return nil
}

func (o *OllamaClient) Chat(ctx context.Context, msg string) ([]Chat, error) {
	//TODO implement me
	panic("implement me")
}

func (o *OllamaClient) pullModel(ctx context.Context, model string) error {
	u, err := url.Parse(o.hostURL)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "ollama", "pull", model)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("OLLAMA_HOST=%s", u.Host))

	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
func (o *OllamaClient) FunctionCalls(ctx context.Context, msg string, chatList ...Chat) ([]Chat, error) {
	if msg == "" {
		return nil, fmt.Errorf("empty msg")
	}
	chatList = append(chatList, Chat{
		Role:    RoleUser,
		Message: msg,
	})
	p, err := getContext(o.externalFunctions, chatList...)
	if err != nil {
		return nil, err
	}

	u := o.hostURL + "/api/generate"
	r := Request{
		Model:  o.model,
		Prompt: p,
	}
	requestBody, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	text := ""
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		r := Response{}
		t := scanner.Text()
		err = json.Unmarshal([]byte(t), &r)
		if err != nil {
			return nil, err
		}
		text += r.Response
		if r.Error != "" {
			if strings.Contains(r.Error, "not found") {
				err = o.pullModel(ctx, o.model)
				if err != nil {
					resp, err = o.client.Do(req)
					if err != nil {
						return nil, err
					}
				}
				continue
			} else {
				return nil, errors.New(r.Error)
			}

		}
	}
	chat, err := GetJSON(text)
	if err != nil {
		return nil, err
	}
	chat.Role = RoleAssistant
	chatList = append(chatList, *chat)
	if f, ok := o.externalFunctionsMap[chat.Tool.ExternalFunctions.Name]; ok {
		output := map[string]interface{}{}
		for _, v := range chat.Tool.ExternalFunctions.Param {
			output[v.Name] = v.Value
			if output[v.Name] == nil {
				output[v.Name] = v.Example
			}
		}
		callResponse, err := f.Call(ctx, output)
		if err != nil {
			return nil, fmt.Errorf("failed to call function(%s): %w", chat.Tool.ExternalFunctions.Name, err)
		}
		chat.Tool.Response = callResponse
		chatList = append(chatList, Chat{
			Role:     RoleSystem,
			Tool:     chat.Tool,
			ChatType: ChatFunctionCallResponse,
		})
	}

	return chatList, nil
}

func (o *OllamaClient) AddFunctions(efList ...*ExternalFunctions) {
	for _, ef := range efList {
		if _, ok := o.externalFunctionsMap[ef.Name]; ok {
			continue
		}
		o.externalFunctions = append(o.externalFunctions, ef)
		o.externalFunctionsMap[ef.Name] = ef

	}

}
