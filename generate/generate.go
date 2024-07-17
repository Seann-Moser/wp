package generate

import (
	"context"
)

type ChatType string

const ChatTypeMessage = ChatType("message")
const ChatTypeTool = ChatType("tool")
const ChatFunctionCall = ChatType("function_call")
const ChatFunctionCallResponse = ChatType("function_call_response")

var ChatTypes = []ChatType{ChatTypeTool, ChatTypeMessage, ChatFunctionCall}

type Role string

const RoleSystem = Role("system")
const RoleUser = Role("user")
const RoleAssistant = Role("assistant")

var Roles = []Role{RoleSystem, RoleUser, RoleAssistant}

type Generator interface {
	GenerateParser(ctx context.Context, url string) error
	Chat(ctx context.Context, msg string) ([]Chat, error)
	FunctionCalls(ctx context.Context, msg string, chatList ...Chat) ([]Chat, error)
	AddFunctions(efList ...*ExternalFunctions)
}

type Chat struct {
	Role     Role     `json:"role"`
	Message  string   `json:"message"`
	Tool     Tool     `json:"tool"`
	ChatType ChatType `json:"chat_type"`
}

type Tool struct {
	ExternalFunctions ExternalFunctions `json:"external_functions"`
	Response          interface{}       `json:"response"`
}

type ExternalFunctions struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Param    []ParamDetails                                                               `json:"param"`
	Call     func(ctx context.Context, param map[string]interface{}) (interface{}, error) `json:"-"`
	Response interface{}                                                                  `json:"response,omitempty"`
}

type ParamDetails struct {
	Name           string        `json:"name"`
	Type           string        `json:"type"`
	Description    string        `json:"description"`
	PossibleValues []interface{} `json:"possible_values,omitempty"`
	Value          interface{}   `json:"value"`
	Example        interface{}   `json:"example,omitempty"`
}
