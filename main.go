package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"ukeuri/config"

	"github.com/bwmarrin/discordgo"
)

func start(listenerSession *discordgo.Session, speakerSession *discordgo.Session) error {
	config, err := config.GetConfig()
	if err != nil {
		return err
	}
	listenerSession.Token = "Bot " + config.ListenerBotID
	speakerSession.Token = "Bot " + config.SpeakerBotID

	err = listenerSession.Open()
	if err != nil {
		log.Println("Failed : Start Listener Bot")
		return err
	}

	err = speakerSession.Open()
	if err != nil {
		log.Println("Failed : Start Speaker Bot")
		return err
	}

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
	listenerSession, err := discordgo.New()
	if err != nil {
		panic(err)
	}

	speakerSession, err := discordgo.New()
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
