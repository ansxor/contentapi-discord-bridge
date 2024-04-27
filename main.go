package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	bot "github.com/ansxor/contentapi-discord-bridge/bot"
	"github.com/ansxor/contentapi-discord-bridge/contentapi"
	"github.com/ansxor/contentapi-discord-bridge/markup"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var contentapi_domain string
var contentapi_token string
var markupService *markup.MarkupService

func ContentApiConnection(session *discordgo.Session, db *sql.DB) {
	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	userId, err := contentapi.GetUserId(contentapi_domain, contentapi_token)
	if err != nil {
		panic(err)
	}

	// this is a flag to make sure that the program fails if the connection fails
	// if we haven't connected before. this way we know whether to retry or not.
	connectedBefore := false

restart_connection:
	connection, resp, err := dialer.Dial(fmt.Sprintf("wss://%s/api/live/ws?token=%s", contentapi_domain, contentapi_token), nil)

	if err != nil {
		if resp != nil {
			fmt.Println(resp.Status)

			bodyBytes, bodyErr := io.ReadAll(resp.Body)
			if bodyErr != nil {
				panic(bodyErr)
			}
			fmt.Println("Body:", string(bodyBytes))
		}
		if connectedBefore {
			log.Default().Println("Reconnection failed. Waiting 10 seconds before trying again: ", err)
			time.Sleep(10 * time.Second)
			goto restart_connection
		}
		log.Fatal("Cannot connect to ContentAPI. Is there an issue with configuration?: ", err)
		panic(err)
	}

	defer connection.Close()
	connectedBefore = true

	log.Default().Println("Connected to ContentAPI")

	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Default().Println("WebSocket closed, reconnecting:", err)
				goto restart_connection
			}
			log.Default().Println("There was an unexpected error reading from the WebSocket:", err)
			continue
		}

		messageEvents, err := contentapi.ParseMessageEvents(string(message))
		if err != nil {
			log.Default().Println(err)
			continue
		}

		for i := range messageEvents {
			event := messageEvents[i]
			// filter all messages sent by the Discord bot
			// NOTE: should this be changed so that this only happens on MessageCreate events?
			if event.User.Id == userId {
				continue
			}

			fmt.Println("Message Event:" + event.Message.Text)
			if event.State == contentapi.MessageCreated {
				channels, err := bot.GetDiscordChannelsFromContentApiRoom(db, event.ContentId)

				if err != nil {
					log.Default().Println(err)
					continue
				}

				for j := range channels {
					channel := channels[j]
					msg, err := bot.WriteDiscordMessage(session, markupService, contentapi_domain, channel, event)
					if err != nil {
						log.Default().Println(err)
						continue
					}

					err = bot.StoreWebhookMessage(db, *msg)
					if err != nil {
						log.Default().Println("Failed storing webhook message:", err)
						continue
					}
				}
			} else if event.State == contentapi.MessageUpdated {
				webhookMessages, err := bot.GetWebhookMessagesForContentApiMessage(db, event.Message.Id)
				if err != nil {
					log.Default().Println(err)
					continue
				}

				for _, webhookMessage := range webhookMessages {
					err := bot.EditDiscordMessage(session, markupService, contentapi_domain, event, webhookMessage)
					if err != nil {
						log.Default().Println(err)
						continue
					}
				}
			} else if event.State == contentapi.MessageDeleted {
				webhookMessages, err := bot.GetWebhookMessagesForContentApiMessage(db, event.Message.Id)
				if err != nil {
					log.Default().Println(err)
					continue
				}

				for _, webhookMessage := range webhookMessages {
					bot.DeleteDiscordMessage(session, contentapi_domain, webhookMessage)
				}

				err = bot.RemoveWebhookMessagesForContentApiMessage(db, event.Message.Id)
				if err != nil {
					log.Default().Println(err)
					continue
				}
			}
		}
	}
}

func GetUsername(member *discordgo.Member) string {
	if member == nil {
		return "Unknown"
	}

	if name := member.Nick; name != "" {
		return name
	} else if member.User != nil {
		if name := member.User.GlobalName; name != "" {
			return name
		}
		return member.User.Username
	}

	return "Unknown"
}

func MessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}

	if strings.HasPrefix(message.Content, "[bind]") {
		params := strings.Split(message.Content, " ")
		if len(params) < 2 {
			return
		}

		room, err := strconv.Atoi(params[1])
		if err != nil {
			log.Default().Println(err)
			return
		}

		_, err = bot.AddChannelPair(db, message.ChannelID, room)
		if err != nil {
			log.Default().Println(err)
			return
		}

		_, err = session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
			Content: fmt.Sprintf("Bound to room %d", room),
			Reference: &discordgo.MessageReference{
				ChannelID: message.ChannelID,
				MessageID: message.ID,
			},
		})
		if err != nil {
			log.Default().Println(err)
			return
		}

		return
	} else if message.Content == "[unbind]" {
		// TODO: we should have a check whether any channels are bound to this room
		err := bot.DisassociateChannel(db, message.ChannelID)
		if err != nil {
			log.Default().Println(err)
			return
		}

		_, err = session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
			Content: "Channel unbound from any references.",
			Reference: &discordgo.MessageReference{
				ChannelID: message.ChannelID,
				MessageID: message.ID,
			},
		})
		if err != nil {
			log.Default().Println(err)
			return
		}

		return
	}

	room, err := bot.GetContentApiRoomFromDiscordChannel(db, message.ChannelID)

	if err != nil {
		fmt.Println(err)
		return
	}

	if room == nil {
		return
	}

	hash, err := bot.FetchAvatarFromUser(db, contentapi_domain, contentapi_token, message.Author)
	if err != nil {
		fmt.Println(err)
		return
	}

	member, err := session.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Default().Println("There was an error getting the Guild Member. Make sure you have the right permissions for your bot: ", err)
		return
	}

	name := GetUsername(member)
	content, err := markupService.DiscordMarkdownToMarkup(message.Content)
	if err != nil {
		log.Default().Println(err)
		return
	}

	for _, attachment := range message.Attachments {
		// attach a newline only if the content is empty so there's isn't a blank line
		if content != "" {
			content += "\n"
		}
		attachment_text := "!" + attachment.URL
		if strings.HasPrefix(attachment.Filename, "SPOILER_") {
			attachment_text = "{#spoiler " + attachment_text + "}"
		}
		content += attachment_text
	}

	id, err := contentapi.ContentApiWriteMessage(contentapi_domain, contentapi_token, *room, content, name, *hash, "12y")

	if err != nil {
		log.Default().Println("There was an error writing the message to ContentAPI:", err)
		return
	}

	contentapiMessage := bot.ContentApiMessageData{
		DiscordMessageId:    message.ID,
		ContentApiMessageId: id,
		ContentApiRoomId:    *room,
	}

	bot.StoreContentApiMessage(db, contentapiMessage)
}

func MessageEdit(session *discordgo.Session, message *discordgo.MessageUpdate) {
	// For some reason, there seems to be cases when editing the messages results
	// in these being nil
	if message == nil || message.Author == nil {
		return
	}

	if message.Author.Bot {
		return
	}

	contentapi_message, err := bot.GetContentApiMessageForDiscordMessage(db, message.ID)

	if err != nil {
		fmt.Println(err)
		return
	}

	if contentapi_message == nil {
		return
	}

	hash, err := bot.FetchAvatarFromUser(db, contentapi_domain, contentapi_token, message.Author)
	if err != nil {
		fmt.Println(err)
		return
	}

	member, err := session.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Default().Println("There was an error getting the Guild Member. Make sure you have the right permissions for your bot: ", err)
		return
	}

	name := GetUsername(member)
	content, err := markupService.DiscordMarkdownToMarkup(message.Content)
	if err != nil {
		log.Default().Println(err)
		return
	}

	for _, attachment := range message.Attachments {
		// attach a newline only if the content is empty so there's isn't a blank line
		if content != "" {
			content += "\n"
		}
		content += "!" + attachment.URL
	}

	err = contentapi.ContentApiEditMessage(contentapi_domain, contentapi_token, contentapi_message.ContentApiMessageId, contentapi_message.ContentApiRoomId, content, name, *hash, "12y")

	if err != nil {
		log.Default().Println("There was an error editing the message on ContentAPI:", err)
		return
	}
}

func MessageDelete(session *discordgo.Session, message *discordgo.MessageDelete) {
	contentapi_message, err := bot.GetContentApiMessageForDiscordMessage(db, message.ID)

	if err != nil {
		fmt.Println(err)
		return
	}

	if contentapi_message == nil {
		return
	}

	err = contentapi.ContentApiDeleteMessage(contentapi_domain, contentapi_token, contentapi_message.ContentApiMessageId)

	if err != nil {
		log.Default().Println("There was an error deleting the message on ContentAPI:", err)
		return
	}
}

func main() {
	contentapi_domain = os.Getenv("CONTENTAPI_DOMAIN")
	contentapi_token = os.Getenv("CONTENTAPI_TOKEN")
	markupService = &markup.MarkupService{
		Domain: os.Getenv("MARKUP_SERVICE_DOMAIN"),
	}

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	intents := discordgo.MakeIntent(
		discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildMembers |
			discordgo.IntentsAllWithoutPrivileged,
	)

	err = dg.Open()
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	dg.Identify.Intents = intents

	dg.AddHandler(MessageCreate)
	dg.AddHandler(MessageEdit)
	dg.AddHandler(MessageDelete)

	db, err = sql.Open("sqlite3", "file:"+os.Getenv("DB_FILE")+"?cache=shared")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	err = bot.InitChannelStore(db)
	if err != nil {
		panic(err)
	}

	err = bot.InitAvatarStore(db)
	if err != nil {
		panic(err)
	}

	err = bot.InitWebhookMessageStore(db)
	if err != nil {
		panic(err)
	}

	err = bot.InitContentApiMessageStore(db)
	if err != nil {
		panic(err)
	}

	go ContentApiConnection(dg, db)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
