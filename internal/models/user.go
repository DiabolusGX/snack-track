package models

type User struct {
	UserId    string      `bson:"user_id" json:"user_id"`
	ChannelId string      `bson:"channel_id" json:"channel_id"`
	TeamId    string      `bson:"team_id" json:"team_id"`
	Schedule  []*Schedule `bson:"schedule" json:"schedule"`
}

type Schedule struct {
	From string `bson:"from" json:"from"`
	To   string `bson:"to" json:"to"`
}
