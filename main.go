package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/timmykkk/maas-go-sdk/sdk"
	volc_maas "github.com/volcengine/volc-sdk-golang/service/maas"
	volc_api "github.com/volcengine/volc-sdk-golang/service/maas/models/api"
)

func main() {
	maasClient, err := sdk.NewMaasClient("xxx")
	if err != nil {
		fmt.Println(err)
		return
	}

	req := &openai.ChatCompletionRequest{
		Model: "skylark2-pro-4k",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "天为什么这么蓝？",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "因为有你",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "花儿为什么这么香？",
			},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
		TopP:        0.9,
	}
	resp, err := maasClient.Chat(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(mustMarshalJson(resp))
}

func mustMarshalJson(v interface{}) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func ChatVolc(r *volc_maas.MaaS, req *volc_api.ChatReq) {
	got, status, err := r.Chat(req)
	if err != nil {
		errVal := &volc_api.Error{}
		if errors.As(err, &errVal) { // the returned error always type of *api.Error
			fmt.Printf("meet maas error=%v, status=%d\n", errVal, status)
		}
		return
	}
	fmt.Println("chat answer", mustMarshalJson(got))
}

func StreamChatVolc(r *volc_maas.MaaS, req *volc_api.ChatReq) {
	ch, err := r.StreamChat(req)

	if err != nil {
		errVal := &volc_api.Error{}
		if errors.As(err, &errVal) { // the returned error always type of *api.Error
			fmt.Println("meet maas error", errVal.Error())
		}
		return
	}

	for resp := range ch {
		if resp.Error != nil {
			// it is possible that error occurs during response processing
			fmt.Println(mustMarshalJson(resp.Error))
			return
		}
		fmt.Println(mustMarshalJson(resp))
		// last response may contain `usage`
		if resp.Usage != nil {
			// last message, will return full response including usage, role, finish_reason, etc.
			fmt.Println(mustMarshalJson(resp.Usage))
		}
	}
}
