package main

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"ukeuri/config"

	"github.com/bwmarrin/discordgo"
)

const (
	VOICE_CHANNEL_TYPE = 2
)

type Status struct {
	Connection *discordgo.VoiceConnection
}

func (status Status) onMessageReceived(session *discordgo.Session, event *discordgo.MessageCreate) {
	// mentionされたときのみ処理を通す
	me, err := session.User("@me")
	if err != nil {
		return
	}
	if len(event.Mentions) == 0 {
		return
	}
	for i, user := range event.Mentions {
		if user.ID == me.ID {
			break
		}
		if i+1 == len(event.Mentions) {
			return
		}
	}

	config, err := config.GetConfig()
	if err != nil {
		log.Fatalln("configの取得に失敗")
		return
	}

	botIDPattern := regexp.MustCompile(`<@\!\d*>`)
	str := botIDPattern.ReplaceAllString(event.Content, "")
	str = strings.TrimSpace(str)
	command := strings.Split(str, " ")

	// Help
	if command[0] == config.HelpCommand {
		message := &discordgo.MessageEmbed{
			Title: "ヘルプ",
			Fields: []*discordgo.MessageEmbedField{{
				Name:   "VCへ参加",
				Value:  "`<Mention> " + config.JoinCommand + " <VoiceChannelName>`",
				Inline: true,
			}, {
				Name:   "VCから退出",
				Value:  "`<Mention>" + config.LeaveCommand + "`",
				Inline: true,
			}},
		}

		_, err = session.ChannelMessageSendEmbed(event.ChannelID, message)
		if err != nil {
			log.Fatalln("ヘルプメッセージの送信に失敗")
		}
		return
	}

	// Join VC
	if command[0] == config.JoinCommand {
		if len(command) == 1 {
			return
		}
		channelName := command[1]

		guild, err := session.State.Guild(event.GuildID)
		if err != nil {
			log.Fatalln(err)
			return
		}

		for _, channel := range guild.Channels {
			if channel.Name == channelName && channel.Type == VOICE_CHANNEL_TYPE {
				status.Connection, err = session.ChannelVoiceJoin(guild.ID, channel.ID, false, false)
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		}
	}
}

func start(listenerSession *discordgo.Session, speakerSession *discordgo.Session) error {
	err := listenerSession.Open()
	if err != nil {
		log.Println("Failed : Start Listener Bot")
		return err
	}

	err = speakerSession.Open()
	if err != nil {
		log.Println("Failed : Start Speaker Bot")
		return err
	}

	listenerStatus := Status{}
	listenerSession.AddHandler(listenerStatus.onMessageReceived)
	speakerStatus := Status{}
	speakerSession.AddHandler(speakerStatus.onMessageReceived)

	return nil
}

func stop(listenerSession *discordgo.Session, speakerSession *discordgo.Session) error {
	err := listenerSession.Close()
	if err != nil {
		return err
	}

	err = speakerSession.Close()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	listenerSession, err := discordgo.New("Bot " + config.ListenerBotID)
	if err != nil {
		panic(err)
	}

	speakerSession, err := discordgo.New("Bot " + config.SpeakerBotID)
	if err != nil {
		panic(err)
	}

	err = start(listenerSession, speakerSession)
	if err != nil {
		panic(err)
	}
	defer stop(listenerSession, speakerSession)

	// 終了を待機
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	select {
	case <-signalChan:
		return
	}
}
