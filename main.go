package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := DISCORD_BOT_TOKEN
	if token == "" {
		log.Fatal("No DISCORD_BOT_TOKEN provided")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	discordconfig := DiscordConfig{
		ForumChannelId:   "1103783859440603198",
		LoggingChannelId: "1046718145538314241",
	}

	var models []Model
	models, err = ListModels()
	// fmt.Printf("Models(%v)", models)

	dg.AddHandler(AssignForum(&discordconfig))
	dg.AddHandler(AssignLoggingChannel(&discordconfig))
	dg.AddHandler(ChatCompletionForum(&discordconfig))

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	defer dg.Close()

	select {}
}
