package hypebot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
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
					Description: "Timestamp to begin song in 9m45s format.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "duration",
					Description: "How long to play your hype song from start time in seconds. (Default: 3, Max: 15)",
					Required:    true,
				},
			},
		},
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	msg                string
)

func (hb *HypeBot) clearCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	exists := hb.userStore.Find(i.GuildID, i.Member.User.ID)

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
	opts := i.ApplicationCommandData().Options

	url, start, duration, err := sanitizeSetCommand(opts)
	log.Printf("%s:%s set %s • [%s, %s, %s] \n", i.Member.User.Username, i.Member.User.ID, i.GuildID, url, start, duration)
	if err != nil {
		msg = err.Error()
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	filePath, err := hb.downloadVideo(url, start, duration)
	if err != nil {
		msg = err.Error()
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
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

	start_time, err := time.ParseDuration(args[1].StringValue())
	if err != nil {
		return "", "", "", fmt.Errorf("`%v`", err.Error())
	}

	dur, err := strconv.ParseFloat(args[2].StringValue(), 64)
	if err != nil {
		return "", "", "", fmt.Errorf("`%v` is not a valid time.", dur)
	}

	if dur <= 0 || dur > 15 {
		return "", "", "", fmt.Errorf("duration `%v` can't be less than **`0`** or greater than **`15`**.", dur)
	}

	convert_start_time := strconv.FormatFloat(start_time.Seconds(), 'f', -1, 64)

	return args[0].StringValue(), convert_start_time, args[2].StringValue(), nil
}
