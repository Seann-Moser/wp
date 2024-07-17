package generate

import (
	"encoding/json"
	"fmt"
	"strings"
)

func GetJSON(msg string) (*Chat, error) {
	startIndex := strings.Index(msg, "{")
	if startIndex < 0 {
		return nil, fmt.Errorf("invalid json format")
	}
	lastIndex := strings.LastIndex(msg, "}")
	if lastIndex < 0 {
		return nil, fmt.Errorf("invalid json format")
	}
	msg = msg[startIndex : lastIndex+1]
	c := Chat{}
	err := json.Unmarshal([]byte(msg), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func getContext(tools []*ExternalFunctions, chatList ...Chat) (string, error) {
	prompts := `
RULES:
If tools are present, in order to call tools you will need to return with a "chat" json object with the name of the function provided.
- All values in JSON objects can be replaces or appended too. 
- NEVER change the keys of defined objects
- There can be an example provided for function call parameters
- You can only call tools that are defined under "TOOLS:"
- "chat_type"" has be one of the following: "%s"

Chat:
%s
`

	c, err := json.MarshalIndent(chatList, "", "  ")
	if err != nil {
		return "", err
	}
	//roleTypeList, err := json.Marshal(Roles)
	//if err != nil {
	//	return "", err
	//}
	chatTypeList, err := json.Marshal(ChatTypes)
	if err != nil {
		return "", err
	}
	prompts = fmt.Sprintf(prompts, string(chatTypeList), string(c))
	if len(tools) > 0 {
		t, err := json.MarshalIndent(tools, "", "  ")
		if err != nil {
			return "", err
		}
		prompts = fmt.Sprintf("%s\n TOOLS:\n%s", prompts, string(t))
	}

	chatObject, err := json.MarshalIndent(Chat{
		Role:    "",
		Message: "",
		Tool:    Tool{},
	}, "", "  ")
	if err != nil {
		return "", err
	}

	prompts = fmt.Sprintf(`%s\n\nONLY Respond with JSON objects formatted as, the values in this object can change but they keys should stay the same:\n%s\n`, prompts, string(chatObject))
	return prompts, nil
}
