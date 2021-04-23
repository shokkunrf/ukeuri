package main

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"ukeuri/config"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

const (
	VOICE_CHANNEL_TYPE = 2
)

type VoiceChat struct {
	Connection    *discordgo.VoiceConnection
	ReceivedVoice chan *discordgo.Packet
	Communicate   func(*discordgo.VoiceConnection, chan *discordgo.Packet)
}

func listen(vc *discordgo.VoiceConnection, recv chan *discordgo.Packet) {
	// VC受信
	go dgvoice.ReceivePCM(vc, recv)
}

func speak(vc *discordgo.VoiceConnection, recv chan *discordgo.Packet) {
	// VC送信
	send := make(chan []int16, 2)
	go dgvoice.SendPCM(vc, send)

	vc.Speaking(true)
	defer vc.Speaking(false)

	for {
		p, ok := <-recv
		if !ok {
			return
		}

		send <- p.PCM
	}
}

func (voiceChat *VoiceChat) receiveMessage(session *discordgo.Session, event *discordgo.MessageCreate) {
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

	switch command[0] {
	case config.HelpCommand:
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

	case config.JoinCommand:
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
				voiceChat.Connection, err = session.ChannelVoiceJoin(guild.ID, channel.ID, false, false)
				if err != nil {
					log.Fatalln(err)
				}

				voiceChat.Communicate(voiceChat.Connection, voiceChat.ReceivedVoice)
				return
			}
		}

	case config.LeaveCommand:
		if voiceChat.Connection == nil {
			return
		}

		err = voiceChat.Connection.Disconnect()
		if err != nil {
			log.Fatalln(err)
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

	recv := make(chan *discordgo.Packet, 2)

	listenerVoiceChat := VoiceChat{
		Communicate:   listen,
		ReceivedVoice: recv,
	}
	listenerSession.AddHandler(listenerVoiceChat.receiveMessage)

	speakerVoiceChat := VoiceChat{
		Communicate:   speak,
		ReceivedVoice: recv,
	}
	speakerSession.AddHandler(speakerVoiceChat.receiveMessage)

	return nil
}

func stop(listenerSession *discordgo.Session, speakerSession *discordgo.Session) error {
	for _, connection := range listenerSession.VoiceConnections {
		err := connection.Disconnect()
		if err != nil {
			return err
		}
	}

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
