package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	anthropicSDK "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/invopop/jsonschema"
	"github.com/mattermost/mattermost-plugin-ai/server/ai"
	"github.com/mattermost/mattermost-plugin-ai/server/metrics"
)

const (
	DefaultMaxTokens       = 4096
	MaxToolResolutionDepth = 10
)

type messageState struct {
	posts       []ai.Post
	toolResults []anthropicSDK.ContentBlockParamUnion
	output      chan<- string
	errChan     chan<- error
}

type Anthropic struct {
	client         *anthropicSDK.Client
	defaultModel   string
	tokenLimit     int
	metricsService metrics.LLMetrics
}

func New(llmService ai.ServiceConfig, httpClient *http.Client, metricsService metrics.LLMetrics) *Anthropic {
	client := anthropicSDK.NewClient(
		option.WithAPIKey(llmService.APIKey),
		option.WithHTTPClient(httpClient),
	)

	return &Anthropic{
		client:         client,
		defaultModel:   llmService.DefaultModel,
		tokenLimit:     llmService.TokenLimit,
		metricsService: metricsService,
	}
}

// isValidImageType checks if the MIME type is supported by the Anthropic API
func isValidImageType(mimeType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[mimeType]
}

// conversationToMessages creates a system prompt and a slice of input messages from a bot conversation.
func conversationToMessages(state messageState) (string, []anthropicSDK.MessageParam) {
	systemMessage := ""
	messages := make([]anthropicSDK.MessageParam, 0, len(state.posts))

	var currentBlocks []anthropicSDK.ContentBlockParamUnion
	var currentRole anthropicSDK.MessageParamRole

	flushCurrentMessage := func() {
		if len(currentBlocks) > 0 {
			messages = append(messages, anthropicSDK.MessageParam{
				Role:    anthropicSDK.F(currentRole),
				Content: anthropicSDK.F(currentBlocks),
			})
			currentBlocks = nil
		}
	}

	for _, post := range state.posts {
		switch post.Role {
		case ai.PostRoleSystem:
			systemMessage += post.Message
			continue
		case ai.PostRoleBot:
			if currentRole != "assistant" {
				flushCurrentMessage()
				currentRole = "assistant"
			}
		case ai.PostRoleUser:
			if currentRole != "user" {
				flushCurrentMessage()
				currentRole = "user"
			}
		default:
			continue
		}

		if post.Message != "" {
			textBlock := anthropicSDK.TextBlockParam{
				Type: anthropicSDK.F(anthropicSDK.TextBlockParamTypeText),
				Text: anthropicSDK.F(post.Message),
			}
			currentBlocks = append(currentBlocks, textBlock)
		}

		for _, file := range post.Files {
			if !isValidImageType(file.MimeType) {
				textBlock := anthropicSDK.TextBlockParam{
					Type: anthropicSDK.F(anthropicSDK.TextBlockParamTypeText),
					Text: anthropicSDK.F(fmt.Sprintf("[Unsupported image type: %s]", file.MimeType)),
				}
				currentBlocks = append(currentBlocks, textBlock)
				continue
			}

			data, err := io.ReadAll(file.Reader)
			if err != nil {
				textBlock := anthropicSDK.TextBlockParam{
					Type: anthropicSDK.F(anthropicSDK.TextBlockParamTypeText),
					Text: anthropicSDK.F("[Error reading image data]"),
				}
				currentBlocks = append(currentBlocks, textBlock)
				continue
			}

			imageBlock := anthropicSDK.ImageBlockParam{
				Type: anthropicSDK.F(anthropicSDK.ImageBlockParamTypeImage),
				Source: anthropicSDK.F(anthropicSDK.ImageBlockParamSource{
					Type:      anthropicSDK.F(anthropicSDK.ImageBlockParamSourceTypeBase64),
					MediaType: anthropicSDK.F(anthropicSDK.ImageBlockParamSourceMediaType(file.MimeType)),
					Data:      anthropicSDK.F(base64.StdEncoding.EncodeToString(data)),
				}),
			}
			currentBlocks = append(currentBlocks, imageBlock)
		}
	}

	// Add tool results if this is a user message continuation
	if len(state.toolResults) > 0 {
		if currentRole != "user" {
			flushCurrentMessage()
			currentRole = "user"
		}
		currentBlocks = append(currentBlocks, state.toolResults...)
	}

	flushCurrentMessage()
	return systemMessage, messages
}

func (a *Anthropic) GetDefaultConfig() ai.LLMConfig {
	return ai.LLMConfig{
		Model:              a.defaultModel,
		MaxGeneratedTokens: DefaultMaxTokens,
	}
}

func (a *Anthropic) createConfig(opts []ai.LanguageModelOption) ai.LLMConfig {
	cfg := a.GetDefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func (a *Anthropic) createCompletionRequest(conversation ai.BotConversation, opts []ai.LanguageModelOption) anthropicSDK.MessageNewParams {
	system, messages := conversationToMessages(messageState{posts: conversation.Posts})
	cfg := a.createConfig(opts)
	return anthropicSDK.MessageNewParams{
		Model:    anthropicSDK.F(cfg.Model),
		Messages: anthropicSDK.F(messages),
		System: anthropicSDK.F([]anthropicSDK.TextBlockParam{{
			Type: anthropicSDK.F(anthropicSDK.TextBlockParamTypeText),
			Text: anthropicSDK.F(system),
		}}),
		MaxTokens: anthropicSDK.F(int64(cfg.MaxGeneratedTokens)),
	}
}

func (a *Anthropic) handleToolResolution(conversation ai.BotConversation, state messageState, depth int, opts ...ai.LanguageModelOption) error {
	if depth >= MaxToolResolutionDepth {
		return fmt.Errorf("max tool resolution depth (%d) exceeded", MaxToolResolutionDepth)
	}

	system, messages := conversationToMessages(state)
	cfg := a.createConfig(opts)

	stream := a.client.Messages.NewStreaming(context.Background(), anthropicSDK.MessageNewParams{
		Model:     anthropicSDK.F(cfg.Model),
		MaxTokens: anthropicSDK.F(int64(cfg.MaxGeneratedTokens)),
		Messages:  anthropicSDK.F(messages),
		System:    anthropicSDK.F([]anthropicSDK.TextBlockParam{anthropicSDK.NewTextBlock(system)}),
		Tools:     anthropicSDK.F(convertTools(conversation.Tools.GetTools())),
	})

	go func() {

		message := anthropicSDK.Message{}
		var toolResults []anthropicSDK.ContentBlockParamUnion

		for stream.Next() {
			event := stream.Current()
			message.Accumulate(event)

			// Stream text content immediately
			switch delta := event.Delta.(type) {
			case anthropicSDK.ContentBlockDeltaEventDelta:
				if delta.Text != "" {
					state.output <- delta.Text
				}
			}
		}

		if err := stream.Err(); err != nil {
			state.errChan <- err
			return
		}

		// Check for tool usage after message is complete
		for _, block := range message.Content {
			if block.Type == anthropicSDK.ContentBlockTypeToolUse {
				// Resolve the tool
				result, err := conversation.Tools.ResolveTool(block.Name, func(args any) error {
					return json.Unmarshal(block.Input, args)
				}, conversation.Context)

				if err != nil {
					state.errChan <- fmt.Errorf("tool resolution error: %w", err)
					return
				}

				toolResults = append(toolResults, anthropicSDK.NewToolResultBlock(block.ID, result, false))
			}
		}

		// If tools were used, continue the conversation with the results
		if len(toolResults) > 0 {
			newState := messageState{
				posts:       state.posts,
				toolResults: toolResults,
			}

			// Recursively handle the continued conversation with same channels
			newState.output = state.output
			newState.errChan = state.errChan
			if err := a.handleToolResolution(conversation, newState, depth+1, opts...); err != nil {
				state.errChan <- err
			}
		}
	}()

	return nil
}

func (a *Anthropic) ChatCompletion(conversation ai.BotConversation, opts ...ai.LanguageModelOption) (*ai.TextStreamResult, error) {
	a.metricsService.IncrementLLMRequests()

	output := make(chan string)
	errChan := make(chan error)

	initialState := messageState{
		posts:       conversation.Posts,
		toolResults: nil,
		output:     output,
		errChan:    errChan,
	}

	go func() {
		defer close(output)
		defer close(errChan)
		
		if err := a.handleToolResolution(conversation, initialState, 0, opts...); err != nil {
			errChan <- err
		}
	}()

	return &ai.TextStreamResult{Stream: output, Err: errChan}, nil
}

func (a *Anthropic) ChatCompletionNoStream(conversation ai.BotConversation, opts ...ai.LanguageModelOption) (string, error) {
	// This could perform better if we didn't use the streaming API here, but the complexity is not worth it.
	result, err := a.ChatCompletion(conversation, opts...)
	if err != nil {
		return "", err
	}
	return result.ReadAll(), nil
}

func (a *Anthropic) CountTokens(text string) int {
	return 0
}

// convertTools converts from ai.Tool to anthropicSDK.Tool format
func convertTools(tools []ai.Tool) []anthropicSDK.ToolParam {
	converted := make([]anthropicSDK.ToolParam, len(tools))
	for i, tool := range tools {
		reflector := jsonschema.Reflector{
			AllowAdditionalProperties: false,
			DoNotReference:            true,
		}
		schema := any(reflector.Reflect(tool.Schema))
		converted[i] = anthropicSDK.ToolParam{
			Name:        anthropicSDK.F(tool.Name),
			Description: anthropicSDK.F(tool.Description),
			InputSchema: anthropicSDK.F(schema),
		}
	}
	return converted
}

func (a *Anthropic) TokenLimit() int {
	if a.tokenLimit > 0 {
		return a.tokenLimit
	}
	return 100000
}
