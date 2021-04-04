package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func main() {
	discord, err := discordgo.New()
	if err != nil {
		fmt.Println("Error logging in")
		fmt.Println(err)
	}
	return
}
