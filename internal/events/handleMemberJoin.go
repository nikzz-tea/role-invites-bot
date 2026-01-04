package events

import (
	"log"
	"role-invites-bot/internal/database"
	"role-invites-bot/internal/handlers"

	"github.com/bwmarrin/discordgo"
)

func init() {
	handlers.RegisterEvent(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		invites, err := s.GuildInvites(m.GuildID)
		if err != nil {
			return
		}

		for _, invite := range invites {
			var record database.Invite

			database.DB.First(&record, "code = ? AND guild_id = ?", invite.Code, m.GuildID)

			if invite.Uses > record.Uses {
				s.GuildMemberRoleAdd(m.GuildID, m.User.ID, record.RoleID)

				database.DB.Model(&record).Update("uses", invite.Uses)

				log.Printf("'%v' role was added to the user '%v'", record.RoleID, m.Member.User.Username)

				if invite.Uses == invite.MaxUses-1 {
					database.DB.Delete(&record)
					s.InviteDelete(invite.Code)
					log.Printf("'%v' invite has been deleted", invite.Code)
				}

				return
			}
		}
	})
}
