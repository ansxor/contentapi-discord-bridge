package bot

// TODO: Need to make test cases for all of these, but it's kinda hard to do

import (
	"github.com/ansxor/contentapi-discord-bridge/contentapi"
	"github.com/bwmarrin/discordgo"
)

const webhookName = "ContentAPI Bridge Webhook"

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

func WriteDiscordMessage(session *discordgo.Session, contentApiDomain string, channelId string, message contentapi.MessageEvent) (*WebhookMessageData, error) {
	webhook, err := FindOrCreateWebhook(session, channelId)
	if err != nil {
		return nil, err
	}

	webhookMessage := &discordgo.WebhookParams{
		Content:   message.Message.Text,
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

func EditDiscordMessages(session *discordgo.Session, contentApiDomain string, message contentapi.MessageEvent, discordMessage []WebhookMessageData) error {
	for _, webhookMessageData := range discordMessage {
		webhookMessage := &discordgo.WebhookEdit{
			Content: &message.Message.Text,
		}

		webhook, err := session.Webhook(webhookMessageData.WebhookId)
		if err != nil {
			return err
		}

		_, err = session.WebhookMessageEdit(webhook.ID, webhook.Token, webhookMessageData.WebhookMessageId, webhookMessage)
		if err != nil {
			return err
		}
	}

	return nil
}
