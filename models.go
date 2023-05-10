package main

type Message struct {
	UserId      int64
	ServiceName string
	Login       string
	Password    string
	AddDate     int64
}

type MessageForDelete struct {
	Id        int
	ChatId    int
	MessageId int
	AddDate   int64
}
