package generate

import (
	"context"
	"encoding/json"
	"github.com/Seann-Moser/wp/source_code"
	"net/http"
	"testing"
)

func TestOllama(t *testing.T) {
	ctx := context.Background()
	source := source_code.NewDirect(http.DefaultClient)
	source.Ping(ctx)
	l := NewOlMA(http.DefaultClient, "http://localhost:8888", OllamaModelDeepSeekCoderV2, source)

	err := l.GenerateParser(ctx, "")
	if err != nil {
		return
	}
}

func TestOllamaFunc(t *testing.T) {
	ctx := context.Background()
	source := source_code.NewDirect(http.DefaultClient)
	source.Ping(ctx)
	l := NewOlMA(http.DefaultClient, "http://localhost:8888", OllamaModelDeepSeekCoderV2, source)

	funList := []*ExternalFunctions{
		{
			Name:        "GetURLSourceCode",
			Description: "returns url source code",
			Param: []ParamDetails{
				{
					Name:        "url",
					Type:        "string",
					Description: "the url to get source code for",
					Example:     "https://example.com/test/",
				},
			},
			Call: func(ctx context.Context, param map[string]interface{}) (interface{}, error) {
				data, _, err := source.Get(ctx, param["url"].(string))
				if err != nil {
					return nil, err
				}
				return string(data), nil
			},
		},
	}
	l.AddFunctions(funList...)

	c, err := l.FunctionCalls(ctx, "tell me about this website: https://github.com/Seann-Moser/")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	m, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	println(string(m))
}

func TestJSONExtract(t *testing.T) {
	testData := `
 '''json
	{
		"role": "user",
		"message": "tell me about this website: https://github.com/Seann-Moser/",
		"tool": {
		"external_functions": {
			"name": "GetURLSourceCode",
				"description": "returns url source code",
				"param": [
{
"name": "url",
"type": "string",
"description": "the url to get source code for",
"value": "https://github.com/Seann-Moser/",
"example": "https://example.com/test/"
}
]
},
"response": null
}
}
'''
`
	_, err := GetJSON(testData)
	if err != nil {
		t.Error(err)
	}

}
