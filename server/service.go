package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-ai/server/ai"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

const (
	WhisperAPILimit    = 25 * 1000 * 1000 // 25 MB
	ContextTokenMargin = 1000
	RespondingToProp   = "responding_to"
)

func (p *Plugin) processUserRequestToBot(context ai.ConversationContext) error {
	if context.Post.RootId == "" {
		return p.newConversation(context)
	}

	threadData, err := p.getThreadAndMeta(context.Post.RootId)
	if err != nil {
		return err
	}

	// Cutoff the thread at the post we are responding to avoid races.
	threadData.cutoffAtPostID(context.Post.Id)

	result, err := p.continueConversation(threadData, context)
	if err != nil {
		return err
	}

	responsePost := &model.Post{
		ChannelId: context.Channel.Id,
		RootId:    context.Post.RootId,
	}
	responsePost.AddProp(RespondingToProp, context.Post.Id)
	if err := p.streamResultToNewPost(context.RequestingUser.Id, result, responsePost); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) newConversation(context ai.ConversationContext) error {
	conversation, err := p.prompts.ChatCompletion(ai.PromptDirectMessageQuestion, context)
	if err != nil {
		return err
	}
	conversation.AddUserPost(context.Post)

	result, err := p.getLLM().ChatCompletion(conversation)
	if err != nil {
		return err
	}

	responsePost := &model.Post{
		ChannelId: context.Channel.Id,
		RootId:    context.Post.Id,
	}
	if err := p.streamResultToNewPost(context.RequestingUser.Id, result, responsePost); err != nil {
		return err
	}

	go func() {
		request := "Write a short title for the following request. Include only the title and nothing else, no quotations. Request:\n" + context.Post.Message
		if err := p.generateTitle(request, context.Post.Id); err != nil {
			p.API.LogError("Failed to generate title", "error", err.Error())
			return
		}
	}()

	return nil
}

func (p *Plugin) generateTitle(request string, threadRootID string) error {
	titleRequest := ai.BotConversation{
		Posts: []ai.Post{{Role: ai.PostRoleUser, Message: request}},
	}
	conversationTitle, err := p.getLLM().ChatCompletionNoStream(titleRequest, ai.WithMaxTokens(25))
	if err != nil {
		return errors.Wrap(err, "failed to get title")
	}

	conversationTitle = strings.Trim(conversationTitle, "\n \"'")

	if err := p.saveTitle(threadRootID, conversationTitle); err != nil {
		return errors.Wrap(err, "failed to save title")
	}

	return nil
}

func (p *Plugin) continueConversation(threadData *ThreadData, context ai.ConversationContext) (*ai.TextStreamResult, error) {
	// Special handing for threads started by the bot in response to a summarization request.
	var result *ai.TextStreamResult
	originalThreadID, ok := threadData.Posts[0].GetProp(ThreadIDProp).(string)
	if ok && originalThreadID != "" && threadData.Posts[0].UserId == p.botid {
		threadPost, err := p.pluginAPI.Post.GetPost(originalThreadID)
		if err != nil {
			return nil, err
		}
		threadChannel, err := p.pluginAPI.Channel.Get(threadPost.ChannelId)
		if err != nil {
			return nil, err
		}

		if !p.pluginAPI.User.HasPermissionToChannel(context.Post.UserId, threadChannel.Id, model.PermissionReadChannel) ||
			p.checkUsageRestrictions(context.Post.UserId, threadChannel) != nil {
			responsePost := &model.Post{
				ChannelId: context.Channel.Id,
				RootId:    context.Post.RootId,
				Message:   "Sorry, you no longer have access to the original thread.",
			}
			if err := p.botCreatePost(context.RequestingUser.Id, responsePost); err != nil {
				return nil, err
			}
			return nil, errors.New("user no longer has access to original thread")
		}

		result, err = p.continueThreadConversation(threadData, originalThreadID, context)
		if err != nil {
			return nil, err
		}
	} else {
		prompt, err := p.prompts.ChatCompletion(ai.PromptDirectMessageQuestion, context)
		if err != nil {
			return nil, err
		}
		prompt.AppendConversation(ai.ThreadToBotConversation(p.botid, threadData.Posts))

		result, err = p.getLLM().ChatCompletion(prompt)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (p *Plugin) continueThreadConversation(questionThreadData *ThreadData, originalThreadID string, context ai.ConversationContext) (*ai.TextStreamResult, error) {
	originalThreadData, err := p.getThreadAndMeta(originalThreadID)
	if err != nil {
		return nil, err
	}
	originalThread := formatThread(originalThreadData)

	context.PromptParameters = map[string]string{"Thread": originalThread}
	prompt, err := p.prompts.ChatCompletion(ai.PromptSummarizeThread, context)
	if err != nil {
		return nil, err
	}
	prompt.AppendConversation(ai.ThreadToBotConversation(p.botid, questionThreadData.Posts))

	result, err := p.getLLM().ChatCompletion(prompt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

const ThreadIDProp = "referenced_thread"

// DM the user with a standard message. Run the inferance
func (p *Plugin) summarizePost(postIDToSummarize string, context ai.ConversationContext) (*ai.TextStreamResult, error) {
	threadData, err := p.getThreadAndMeta(postIDToSummarize)
	if err != nil {
		return nil, err
	}

	formattedThread := formatThread(threadData)

	context.PromptParameters = map[string]string{"Thread": formattedThread}
	prompt, err := p.prompts.ChatCompletion(ai.PromptSummarizeThread, context)
	if err != nil {
		return nil, err
	}
	summaryStream, err := p.getLLM().ChatCompletion(prompt)
	if err != nil {
		return nil, err
	}

	return summaryStream, nil
}

func summaryPostMessage(postIDToSummarize string, siteURL string) string {
	return fmt.Sprintf("Sure, I will summarize this thread: %s/_redirect/pl/%s\n", siteURL, postIDToSummarize)
}

func (p *Plugin) makeSummaryPost(postIDToSummarize string) *model.Post {
	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	post := &model.Post{
		Message: summaryPostMessage(postIDToSummarize, *siteURL),
	}
	post.AddProp(ThreadIDProp, postIDToSummarize)

	return post
}

func (p *Plugin) startNewSummaryThread(postIDToSummarize string, context ai.ConversationContext) (*model.Post, error) {
	summaryStream, err := p.summarizePost(postIDToSummarize, context)
	if err != nil {
		return nil, err
	}

	post := p.makeSummaryPost(postIDToSummarize)
	if err := p.streamResultToNewDM(summaryStream, context.RequestingUser.Id, post); err != nil {
		return nil, err
	}

	if err := p.saveTitle(post.Id, "Thread Summary"); err != nil {
		return nil, errors.Wrap(err, "failed to save title")
	}

	return post, nil
}

func (p *Plugin) selectEmoji(postToReact *model.Post, context ai.ConversationContext) error {
	context.PromptParameters = map[string]string{"Message": postToReact.Message}
	prompt, err := p.prompts.ChatCompletion(ai.PromptEmojiSelect, context)
	if err != nil {
		return err
	}

	emojiName, err := p.getLLM().ChatCompletionNoStream(prompt, ai.WithMaxTokens(25))
	if err != nil {
		return err
	}

	// Do some emoji post processing to hopefully make this an actual emoji.
	emojiName = strings.Trim(strings.TrimSpace(emojiName), ":")

	if _, found := model.GetSystemEmojiId(emojiName); !found {
		p.pluginAPI.Post.AddReaction(&model.Reaction{
			EmojiName: "large_red_square",
			UserId:    p.botid,
			PostId:    postToReact.Id,
		})
		return errors.New("LLM returned somthing other than emoji: " + emojiName)
	}

	if err := p.pluginAPI.Post.AddReaction(&model.Reaction{
		EmojiName: emojiName,
		UserId:    p.botid,
		PostId:    postToReact.Id,
	}); err != nil {
		return err
	}

	return nil
}
