package commands

import (
	"fmt"
	"role-invites-bot/internal/handlers"
	"role-invites-bot/internal/models"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type channelSpec struct {
	name string
	kind discordgo.ChannelType
	deny int64
}

var channelSpecs = []channelSpec{
	{"links", discordgo.ChannelTypeGuildText,
		discordgo.PermissionSendMessages |
			discordgo.PermissionCreatePublicThreads |
			discordgo.PermissionCreatePrivateThreads |
			discordgo.PermissionSendMessagesInThreads |
			discordgo.PermissionUseApplicationCommands},
	{"general", discordgo.ChannelTypeGuildText, 0},
	{"vc", discordgo.ChannelTypeGuildVoice, 0},
}

func init() {
	handlers.RegisterCommand(models.CommandObject{
		Name:    "createchat",
		Aliases: []string{"chat", "chatcreate", "makechat"},
		Callback: func(props models.CommandProps) {
			sess, message, args := props.Sess, props.Message, props.Args
			if message.GuildID == "" {
				return
			}
			if len(args) < 2 {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **!createchat <color> <title>**")
				return
			}

			color, err := parseColor(args[0])
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Invalid color** (use a hex value like `#5865f2`)")
				return
			}
			title := strings.Join(args[1:], " ")

			mentionable := true
			role, err := sess.GuildRoleCreate(message.GuildID, &discordgo.RoleParams{
				Name:        title,
				Color:       &color,
				Mentionable: &mentionable,
			})
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to create the role**")
				return
			}
			if err := sess.GuildMemberRoleAdd(message.GuildID, sess.State.User.ID, role.ID); err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to assign the role to the bot**")
				return
			}

			accessOverwrites := []*discordgo.PermissionOverwrite{
				{ID: message.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
				{ID: role.ID, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel},
			}
			category, err := sess.GuildChannelCreateComplex(message.GuildID, discordgo.GuildChannelCreateData{
				Name:                 title,
				Type:                 discordgo.ChannelTypeGuildCategory,
				PermissionOverwrites: accessOverwrites,
			})
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to create the category**")
				return
			}

			for _, spec := range channelSpecs {
				overwrites := []*discordgo.PermissionOverwrite{
					{ID: message.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
					{ID: role.ID, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel, Deny: spec.deny},
				}
				if _, err := sess.GuildChannelCreateComplex(message.GuildID, discordgo.GuildChannelCreateData{
					Name:                 spec.name,
					Type:                 spec.kind,
					ParentID:             category.ID,
					PermissionOverwrites: overwrites,
				}); err != nil {
					sess.ChannelMessageSend(message.ChannelID, fmt.Sprintf("🔴 **Failed to create the channel `%v`**", spec.name))
					return
				}
			}

			if err := sess.GuildMemberRoleRemove(message.GuildID, sess.State.User.ID, role.ID); err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to remove the role from the bot**")
			}

			sess.ChannelMessageSend(
				message.ChannelID,
				fmt.Sprintf("✅ **Created role <@&%v> and category `%v`**", role.ID, title),
			)
		},
	})
}

func parseColor(input string) (int, error) {
	hex := strings.TrimSpace(input)
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.TrimPrefix(hex, "0x")
	if len(hex) != 6 {
		return 0, fmt.Errorf("color must be a 6-digit hex value")
	}
	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, err
	}
	return int(value), nil
}
