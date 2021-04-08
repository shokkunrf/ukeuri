package config

import (
	"log"

	"github.com/usagiga/envs-go"
)

type Config struct {
	ListenerBotID string `envs:"LISTENER_BOT_ID"`
	SpeakerBotID  string `envs:"SPEAKER_BOT_ID"`
	HelpCommand   string
	JoinCommand   string
	LeaveCommand  string
}

func GetConfig() (Config, error) {
	config := &Config{
		HelpCommand:  "help",
		JoinCommand:  "join",
		LeaveCommand: "leave",
	}

	err := envs.Load(config)
	if err != nil {
		log.Fatalf("Can't load config: %+v", err)
	}

	return *config, nil
}
