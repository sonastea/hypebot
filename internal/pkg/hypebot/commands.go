package hypebot

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "clear",
			Description: "Clear your currently set themesong.",
		},
		{
			Name:        "set",
			Description: "Set a themesong given a youtube link and start time in seconds.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "song-url",
					Description: "YouTube URL of the song you want to use.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "start-time",
					Description: "Starting time of the song in seconds.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "duration",
					Description: "Length to play your hype song from the star time.",
					Required:    true,
				},
			},
		},
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	msg                string
)

func (hb *HypeBot) clearCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Processing clear command...",
		},
	})

	exists := users.FindUser(hb.db, i.GuildID, i.Member.User.ID)

	if !exists {
		msg = "You have not set a themesong with HypeBot."
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})

		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})
			return
		}
	} else {
		msg = hb.removeThemesong(i.GuildID, i.Member.User.ID)
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})

		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Something went wrong",
			})
			return
		}
	}
}

func (hb *HypeBot) setCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Processing set command...",
		},
	})

	opts := i.ApplicationCommandData().Options

	url, start, duration, err := sanitizeSetCommand(opts)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: err.Error(),
			},
		})
		return
	}

	filePath, err := hb.downloadVideo(url, start, duration)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: err.Error(),
			},
		})
		return
	}

	msg = hb.setThemesong(filePath, i.GuildID, i.Member.User.ID)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})

	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		})
		return
	}
}

func sanitizeSetCommand(args []*discordgo.ApplicationCommandInteractionDataOption) (string, string, string, error) {
	validUrl, _ := regexp.MatchString(`(?:https?:\/\/)?(?:www\.)?youtu\.?be(?:\.com)?\/?.*(?:watch|embed)?(?:.*v=|v\/|\/)([\w\-_]+)\&?`, args[0].StringValue())
	if !validUrl {
		return "", "", "", fmt.Errorf("`%v` is not a valid url.", args[0].StringValue())
	}

	for i := 1; i <= 2; i++ {
		_, err := strconv.ParseFloat(args[i].StringValue(), 64)
		if err != nil {
			return "", "", "", fmt.Errorf("`%v` is not a valid time.", args[i].StringValue())
		}
	}

	return args[0].StringValue(), args[1].StringValue(), args[2].StringValue(), nil
}
