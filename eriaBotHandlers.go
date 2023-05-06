package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func AssignForum(discordconfig *DiscordConfig) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if strings.HasPrefix(m.Content, "!GPTstart") {
			channelId := strings.TrimPrefix(m.Content, "!GPTstart ")
			channelId = strings.TrimPrefix(channelId, "<#")
			channelId = strings.TrimSuffix(channelId, ">")

			_, convErr := strconv.Atoi(channelId)
			if convErr != nil {
				s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error configuring channel, not a valid channel.")
			}

			channel, err := s.Channel(channelId)
			if err != nil {
				s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error configuring channel: "+err.Error())
				return
			}
			discordconfig.ForumChannelId = channel.ID
			s.ChannelMessageSend(m.ChannelID, "Channel configured: "+"<#"+channelId+">")
		}
	}
}

func AssignLoggingChannel(discordconfig *DiscordConfig) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if strings.HasPrefix(m.Content, "!loggingChannel") {
			channelId := strings.TrimPrefix(m.Content, "!loggingChannel ")
			channelId = strings.TrimPrefix(channelId, "<#")
			channelId = strings.TrimSuffix(channelId, ">")

			_, convErr := strconv.Atoi(channelId)
			if convErr != nil {
				s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error configuring channel, not a valid channel.")
			}

			channel, err := s.Channel(channelId)
			if err != nil {
				s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error configuring channel: "+err.Error())
				return
			}
			discordconfig.LoggingChannelId = channel.ID
			s.ChannelMessageSend(m.ChannelID, "Channel configured: "+"<#"+channelId+">")
		}
	}
}

func ChatCompletionForum(discordconfig *DiscordConfig) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		ch, err := s.Channel(m.ChannelID)

		if err != nil || ch.ParentID != discordconfig.ForumChannelId {
			return
		}

		s.ChannelTyping(m.ChannelID)

		messages, err := s.ChannelMessages(m.ChannelID, MAX_MESSAGE_CHATGPT_SENT*2, m.ID, "", "")
		if err != nil {
			s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error retrieving messages: "+err.Error())
			return
		}

		chatGPTConversation := make([]map[string]string, 0)
		for _, message := range messages {
			role := "user"
			if message.Author.ID == s.State.User.ID {
				role = "assistant"
			}

			newMap := map[string]string{
				"role":    role,
				"content": message.Content,
			}

			chatGPTConversation = append(chatGPTConversation, newMap)
		}

		response, err := ChatGPTChatCompletion(chatGPTConversation, m.Content)
		if err != nil {
			s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error getting response from ChatGPT API: "+err.Error())
			return
		}

		if len(response.Choices) == 0 {
			s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error, no choices in response from ChatGPT API.")
			return
		}

		responseMessage := response.Choices[0].Message.Content
		codeBlocks := strings.Split(responseMessage, "```")
		var splitResponseMessage []string

		for i, block := range codeBlocks {
			toSend := block
			if i%2 != 0 { // If it is a code block
				toSend = "```" + block + "```"
				if len(toSend) >= MAX_DISCORD_MESSAGE_LENGHT { //This is to send a file in case it is bigger than 2000
					splitResponseMessage = append(splitResponseMessage, FILE_MSG_PREFFIX+" ") //HANDLE THIS
					file, err := os.Create("code")
					if err != nil {
						s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error creating file for code: "+err.Error())
						return
					}
					defer file.Close()
					writer := bufio.NewWriter(file)
					_, err = writer.WriteString(toSend)
					if err != nil {
						s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error writing to file: "+err.Error())
						return
					}
					err = writer.Flush()
					if err != nil {
						s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error flushing file: "+err.Error())
						return
					}
				} else {
					splitResponseMessage = append(splitResponseMessage, toSend)
				}
			} else {
				var splitResponseMessageNoBlock []string
				if len(toSend) < MAX_DISCORD_MESSAGE_LENGHT {
					splitResponseMessage = append(splitResponseMessage, toSend)
				} else {
					toSendTmp := toSend
					for len(toSend) >= MAX_DISCORD_MESSAGE_LENGHT {
						start := MAX_DISCORD_MESSAGE_LENGHT * len(splitResponseMessageNoBlock)
						end := MAX_DISCORD_MESSAGE_LENGHT * (len(splitResponseMessageNoBlock) + 1)
						if end > len(toSendTmp) {
							end = len(toSendTmp)
						}
						toSend = toSendTmp[start:end]
						splitResponseMessageNoBlock = append(splitResponseMessageNoBlock, toSend)
					}

					splitResponseMessage = append(splitResponseMessage, splitResponseMessageNoBlock...)
				}
			}
		}

		for _, msg := range splitResponseMessage {
			if strings.HasPrefix(m.Content, FILE_MSG_PREFFIX) {
				fileName := strings.TrimPrefix(m.Content, FILE_MSG_PREFFIX+" ")
				file, err := os.Open(fileName)
				if err != nil {
					file.Close()
					s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error opening code file: "+err.Error())
				}

				_, err = s.ChannelFileSend(m.ChannelID, fileName, file)
				if err != nil {
					file.Close()
					s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error sending code file: "+err.Error())
					return
				}

				file.Close()
				err = os.Remove(fileName)
				if err != nil {
					s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error removing code file: "+err.Error())
					return
				}

			} else {
				_, messageSendError := s.ChannelMessageSend(m.ChannelID, msg)
				if messageSendError != nil {
					s.ChannelMessageSend(discordconfig.LoggingChannelId, "Error printing the ChatGPT response: "+messageSendError.Error())
				}
			}
		}
	}
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!ask") {
		query := strings.TrimPrefix(m.Content, "!ask ")
		response, err := ChatGPTQuery(query)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error getting response from ChatGPT API: "+err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, response.Choices[0].Message.Content)
	} else {
		fmt.Println("No content")
	}
}
