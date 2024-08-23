package models

type User struct {
	UserId     string      `bson:"user_id" json:"user_id"`
	ChannelId  string      `bson:"channel_id" json:"channel_id"`
	TeamDomain string      `bson:"team_domain" json:"team_domain"`
	Schedule   []*Schedule `bson:"schedule" json:"schedule"`
	AddressIds []string    `bson:"address_ids" json:"address_ids"`
}

type Schedule struct {
	From string `bson:"from" json:"from"`
	To   string `bson:"to" json:"to"`
}
