package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"strings"
)

const (
	GPT3Dot5Turbo = "gpt-3.5"
	GPT3          = "gpt-3"
	GPT4          = "gpt-4"
)

var promptMap = map[string]string{
	"type": `You are a technical expert in the public cloud and SaaS. You need to help determine which public cloud or SaaS analysis requirement the question belongs to according to the question I provided, and answer it as described below:
1. For security requirements, please return: security & cloudName
2. For cost analysis requirements, please return: finops & cloudName
3. For architecture optimization requirements, please return: architecture & cloudName

Do not output Answer and line breaks

The question is:
%s`,
	"securityTable": `You are a public cloud and SaaS security expert. What you need to do is to determine which database tables and fields will be used in the analysis question based on the tables and security analysis questions I provided, and return them in the following format
		Table1,Table2,Table3

		Do not output Answer and line breaks
		
The question is:
%s

Tables is:
%s
`,
	"securityColumn": `You are a public cloud and SaaS security expert. What you need to do is to determine which database tables and fields will be used in the analysis question based on the tables and security analysis questions I provided, and return them in the following format

Column1,Column2,Column3

Do not output Answer and line breaks

Do not return duplicate fields

Do not appear fields that do not exist in the original field

The question is:
%s

Table is:
%s

Column is:
%s
`,
	"security": `You are a public cloud and SaaS security expert. I will give you a section of %s's %s configuration information data and security analysis question. Please help me detect whether there is a security vulnerability in this configuration. If there is a security problem, please return the vulnerability title, vulnerability description, repair recommendation for the complete executable steps, security level, Tags of the security compliance framework, and return it in the following example format:
[
	{
		"title":"",
		"description":"",
		"remediation": "",
		"severity": "",
		"tags":[""],
		"resource":"",
	}
]
The configuration is:
%s

The question is:
%s
`,
}

func OpenApiClient(ctx context.Context, sk string, mode string, promptType string, args ...any) (string, error) {
	client := openai.NewClient(sk)
	switch mode {
	case GPT3Dot5Turbo:
		return GPT3Dot5TurboFunc(ctx, client, fmt.Sprintf(promptMap[promptType], args...))
	case GPT3:
		prompt := fmt.Sprintf(promptMap[promptType], args...)
		return GPT3Func(ctx, client, prompt)
	case GPT4:
		return "", errors.New("gpt-4 not support")
	}
	return "", errors.New("mode not found")
}

func GPT3Dot5TurboFunc(ctx context.Context, client *openai.Client, prompt string) (string, error) {
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 512,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	return strings.Trim(resp.Choices[0].Message.Content, "\n"), nil
}

func GPT3Func(ctx context.Context, client *openai.Client, Prompt string) (string, error) {
	req := openai.CompletionRequest{
		Model:     openai.GPT3TextDavinci003,
		MaxTokens: 256,
		Prompt:    Prompt,
	}
	resp, err := client.CreateCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return "", err
	}
	return strings.Trim(resp.Choices[0].Text, "\n"), nil
}

func GPT4Func(ctx context.Context, client *openai.Client, Prompt string) {

}
