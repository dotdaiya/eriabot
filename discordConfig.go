package main

const (
	chatGPTAPITOKEN            = ""
	DISCORD_BOT_TOKEN          = ""
	MAX_MESSAGE_CHATGPT_SENT   = 5
	MAX_DISCORD_MESSAGE_LENGHT = 2000
	MAX_COMPLETION_CHARS       = 4097
	FILE_MSG_PREFFIX           = "[BOT][FILETOSEND]"
)

type DiscordConfig struct {
	ForumChannelId   string `json:"forumChannelId"`
	LoggingChannelId string `json:"loggingChannelId"`
}
