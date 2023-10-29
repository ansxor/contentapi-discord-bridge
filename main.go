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
		log.Fatal("Failed to connect to ContentAPI :/ ", err)
		panic(err)
	}

	defer connection.Close()

	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Default().Println("There was an error reading from the WebSocket:", err)
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

	_, err = contentapi.ContentApiWriteMessage(contentapi_domain, contentapi_token, *room, content, name, *hash, "12y")

	if err != nil {
		log.Default().Println("There was an error writing the message to ContentAPI:", err)
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

	go ContentApiConnection(dg, db)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
