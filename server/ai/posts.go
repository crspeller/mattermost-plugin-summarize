package ai

import (
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type PostRole int

const (
	PostRoleUser PostRole = iota
	PostRoleBot
	PostRoleSystem
)

type Post struct {
	Role    PostRole
	Message string
}

type BotConversation struct {
	Posts []Post
	Tools []Tool
}

func (b *BotConversation) AddUserPost(post *model.Post) {
	b.Posts = append(b.Posts, Post{
		Role:    PostRoleUser,
		Message: post.Message,
	})
}

func (b *BotConversation) AppendConversation(conversation BotConversation) {
	b.Posts = append(b.Posts, conversation.Posts...)
}

func (b *BotConversation) ExtractSystemMessage() string {
	var result strings.Builder
	for _, post := range b.Posts {
		if post.Role == PostRoleSystem {
			result.WriteString(post.Message)
		}
	}
	return result.String()
}

func (b *BotConversation) AddBuiltInTools() {
	b.Tools = append(b.Tools, BuiltInTools...)
}

func GetPostRole(botID string, post *model.Post) PostRole {
	if post.UserId == botID {
		return PostRoleBot
	}
	return PostRoleUser
}

func ThreadToBotConversation(botID string, posts []*model.Post) BotConversation {
	result := BotConversation{
		Posts: make([]Post, 0, len(posts)),
	}

	for _, post := range posts {
		result.Posts = append(result.Posts, Post{
			Role:    GetPostRole(botID, post),
			Message: post.Message,
		})
	}

	return result
}
