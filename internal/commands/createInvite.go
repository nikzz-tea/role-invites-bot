package commands

import (
	"fmt"
	"role-invites-bot/internal/database"
	"role-invites-bot/internal/handlers"
	"role-invites-bot/internal/models"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	handlers.RegisterCommand(models.CommandObject{
		Name:    "createinvite",
		Aliases: []string{"invite", "invitecreate", "makeinvite"},
		Callback: func(props models.CommandProps) {
			sess, message, args := props.Sess, props.Message, props.Args
			if message.GuildID == "" {
				return
			}
			if len(args) < 2 {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **!createinvite <@role> <max uses>**")
				return
			}

			roleID := args[0][3 : len(args[0])-1]
			roles, err := sess.GuildRoles(message.GuildID)
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to fetch roles**")
				return
			}

			var role *discordgo.Role
			for _, r := range roles {
				if r.ID == roleID {
					role = r
					break
				}
			}
			if role == nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Invalid role ID**")
				return
			}

			maxUses, err := strconv.Atoi(args[1])
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Invalid max uses option**")
				return
			}

			expiresAt := time.Now().Add(24 * time.Hour)

			invite, err := sess.ChannelInviteCreate(message.ChannelID, discordgo.Invite{
				MaxUses:   maxUses + 1,
				Unique:    true,
				ExpiresAt: &expiresAt,
			})
			if err != nil {
				sess.ChannelMessageSend(message.ChannelID, "🔴 **Failed to create the invite**")
				return
			}

			database.DB.Create(&database.Invite{
				Code:    invite.Code,
				RoleID:  roleID,
				GuildID: message.GuildID,
				Uses:    invite.Uses,
			})

			sess.ChannelMessageSend(message.ChannelID, fmt.Sprintf("https://discord.gg/%v <@&%v>", invite.Code, roleID))
		},
	})
}
