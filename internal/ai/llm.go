package ai

import (
	"api/internal/initialize"
	"context"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/conversations"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	openaiClient openai.Client
}

func NewClient() *Client {
	return &Client{
		openaiClient: openai.NewClient(
			option.WithAPIKey(initialize.CFG.OpenAIKey),
		),
	}
}

func (c Client) GetAnswerFromAI(ctx context.Context, question string, conversation_id string) (string, error) {
	resp, err := c.openaiClient.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(question),
		},
		Prompt: responses.ResponsePromptParam{
			ID:      conversation_id,
			Version: param.NewOpt(initialize.CFG.PromptVersion),
		},
		Conversation: responses.ResponseNewParamsConversationUnion{
			OfConversationObject: &responses.ResponseConversationParam{
				ID: conversation_id,
			},
		},
		Model: openai.ChatModelGPT4oMini, // используйте актуальную модель
	})
	if err != nil {
		log.Printf("Couldn't get answer from AI: %v", err)
		return "", err
	}
	return resp.OutputText(), nil
}

// NewConversation создаёт новый пустой диалог и возвращает его ID
func (c Client) NewConversation(ctx context.Context) (string, error) {
	conv, err := c.openaiClient.Conversations.New(ctx, conversations.ConversationNewParams{})
	if err != nil {
		log.Printf("Failed to create conversation: %v", err)
		return "", err
	}
	return conv.ID, nil
}
