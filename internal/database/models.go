package database

type Invite struct {
	Code    string `gorm:"primaryKey"`
	RoleID  string
	GuildID string
	Uses    int
}
