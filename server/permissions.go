package main

import (
	"fmt"
	"slices"

	"errors"

	"github.com/mattermost/mattermost-plugin-ai/server/ai"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

var ErrUsageRestriction = errors.New("usage restriction")

func (p *Plugin) checkUsageRestrictions(requestingUserID string, bot *Bot, channel *model.Channel) error {
	if err := p.checkUsageRestrictionsForUser(bot, requestingUserID); err != nil {
		return err
	}

	if err := p.checkUsageRestrictionsForChannel(bot, channel); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) checkUsageRestrictionsForChannel(bot *Bot, channel *model.Channel) error {
	switch bot.cfg.ChannelAccessLevel {
	case ai.ChannelAccessLevelAll:
		return nil
	case ai.ChannelAccessLevelAllow:
		if !slices.Contains(bot.cfg.ChannelIDs, channel.Id) {
			return fmt.Errorf("channel not allowed: %w", ErrUsageRestriction)
		}
		return nil
	case ai.ChannelAccessLevelBlock:
		if slices.Contains(bot.cfg.ChannelIDs, channel.Id) {
			return fmt.Errorf("channel blocked: %w", ErrUsageRestriction)
		}
		return nil
	case ai.ChannelAccessLevelNone:
		return fmt.Errorf("channel usage block for bot: %w", ErrUsageRestriction)
	}

	return fmt.Errorf("unknown channel assistance level")
}

func (p *Plugin) isMemberOfTeam(teamID string, userID string) (bool, error) {
	member, err := p.pluginAPI.Team.GetMember(teamID, userID)
	if errors.Is(err, pluginapi.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return member != nil && member.DeleteAt == 0, nil
}

func (p *Plugin) checkUsageRestrictionsForUser(bot *Bot, requestingUserID string) error {
	switch bot.cfg.UserAccessLevel {
	case ai.UserAccessLevelAll:
		return nil
	case ai.UserAccessLevelAllow:
		// Check direct user allowlist
		if slices.Contains(bot.cfg.UserIDs, requestingUserID) {
			return nil
		}
		// Check team membership
		for _, teamID := range bot.cfg.TeamIDs {
			isMember, err := p.isMemberOfTeam(teamID, requestingUserID)
			if err != nil {
				return err
			}
			if isMember {
				return nil
			}
		}
		return fmt.Errorf("user not allowed: %w", ErrUsageRestriction)
	case ai.UserAccessLevelBlock:
		// Check direct user blocklist
		if slices.Contains(bot.cfg.UserIDs, requestingUserID) {
			return fmt.Errorf("user blocked: %w", ErrUsageRestriction)
		}
		// Check team membership
		for _, teamID := range bot.cfg.TeamIDs {
			isMember, err := p.isMemberOfTeam(teamID, requestingUserID)
			if err != nil {
				return err
			}
			if isMember {
				return fmt.Errorf("user's team blocked: %w", ErrUsageRestriction)
			}
		}
		return nil
	case ai.UserAccessLevelNone:
		return fmt.Errorf("user usage block for bot: %w", ErrUsageRestriction)
	}

	return fmt.Errorf("unknown user assistance level")
}
