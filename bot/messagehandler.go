package bot

// TODO: Need to make test cases for all of these, but it's kinda hard to do

import (
	"regexp"

	"github.com/ansxor/contentapi-discord-bridge/contentapi"
	"github.com/ansxor/contentapi-discord-bridge/markup"
	"github.com/bwmarrin/discordgo"
)

const webhookName = "ContentAPI Bridge Webhook"

func FilterMentions(text string) string {
	pattern := `@(everyone|here)`
	re := regexp.MustCompile(pattern)

	replacementFunc := func(match string) string {
		return "@\u200B" + match[1:]
	}

	return re.ReplaceAllStringFunc(text, replacementFunc)
}

func FindBotWebhook(session *discordgo.Session, channelId string) (*discordgo.Webhook, error) {
	webhooks, err := session.ChannelWebhooks(channelId)
	if err != nil {
		return nil, err
	}

	for _, webhook := range webhooks {
		if webhook.User.ID == session.State.User.ID {
			return webhook, nil
		}
	}

	return nil, nil
}

func FindOrCreateWebhook(session *discordgo.Session, channelId string) (*discordgo.Webhook, error) {
	webhook, err := FindBotWebhook(session, channelId)

	if err != nil {
		return nil, err
	}

	if webhook != nil {
		return webhook, nil
	}

	newWebhook, err := session.WebhookCreate(channelId, webhookName, "")
	if err != nil {
		return nil, err
	}

	return newWebhook, nil
}

func WriteDiscordMessage(session *discordgo.Session, markupService *markup.MarkupService, contentApiDomain string, channelId string, message contentapi.MessageEvent) (*WebhookMessageData, error) {
	webhook, err := FindOrCreateWebhook(session, channelId)
	if err != nil {
		return nil, err
	}

	content, err := markupService.MarkupToDiscordMarkdown(message.Message.Text, message.Message.Markup)
	if err != nil {
		return nil, err
	}

	content = FilterMentions(content)

	webhookMessage := &discordgo.WebhookParams{
		Content:   content,
		AvatarURL: message.User.GetAvatar(contentApiDomain, contentapi.DEFAULT_AVATAR_SIZE),
		Username:  message.User.Username,
	}

	msg, err := session.WebhookExecute(webhook.ID, webhook.Token, true, webhookMessage)
	if err != nil {
		return nil, err
	}

	return &WebhookMessageData{
		WebhookId:               webhook.ID,
		WebhookMessageChannelId: channelId,
		WebhookMessageId:        msg.ID,
		ContentApiMessageId:     message.Message.Id,
	}, nil
}

func EditDiscordMessage(session *discordgo.Session, markupService *markup.MarkupService, contentApiDomain string, message contentapi.MessageEvent, webhookMessage WebhookMessageData) error {
	content, err := markupService.MarkupToDiscordMarkdown(message.Message.Text, message.Message.Markup)
	if err != nil {
		return err
	}

	content = FilterMentions(content)

	webhookMessageData := &discordgo.WebhookEdit{
		Content: &content,
	}

	webhook, err := session.Webhook(webhookMessage.WebhookId)
	if err != nil {
		return err
	}

	_, err = session.WebhookMessageEdit(webhook.ID, webhook.Token, webhookMessage.WebhookMessageId, webhookMessageData)
	if err != nil {
		return err
	}

	return nil
}

func DeleteDiscordMessage(session *discordgo.Session, contentApiDomain string, webhookMessage WebhookMessageData) error {
	webhook, err := session.Webhook(webhookMessage.WebhookId)
	if err != nil {
		return err
	}

	err = session.WebhookMessageDelete(webhook.ID, webhook.Token, webhookMessage.WebhookMessageId)
	if err != nil {
		return err
	}

	return nil
}
