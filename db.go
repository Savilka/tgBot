package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"time"
)

type DB struct {
	db *sql.DB
}

func (DB *DB) init(dbName string) error {
	var err error
	DB.db, err = sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}

	return nil
}

func (DB *DB) close() error {
	err := DB.db.Close()
	if err != nil {
		return err
	}

	return nil
}

func (DB *DB) addMessage(message Message) error {
	stmt, err := DB.db.Prepare("insert into messages (user_id, service_name, login, password, add_date) values (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(message.UserId, message.ServiceName, message.Login, message.Password, message.AddDate)
	if err != nil {
		return err
	}

	return nil
}

func (DB *DB) getMessage(userId int, serviceName string) (Message, error) {
	var message Message
	stmt, err := DB.db.Prepare("select login, password from messages where user_id = ? and service_name = ?")
	if err != nil {
		return message, err
	}

	defer stmt.Close()

	row, err := stmt.Query(userId, serviceName)
	if err != nil {
		return message, err
	}

	defer row.Close()

	if !row.Next() {
		return message, nil
	}
	if err := row.Scan(&message.Login, &message.Password); err != nil { // scan will release the connection
		return message, err
	}

	return message, nil
}

func (DB *DB) deleteMessage(userId int, serviceName string) error {
	stmt, err := DB.db.Prepare("delete from messages where user_id = ? and service_name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	row, err := stmt.Query(userId, serviceName)
	if err != nil {
		return err
	}
	defer row.Close()

	return nil
}

func (DB *DB) addMessageIdToDelete(id int, chatId int64) error {
	stmt, err := DB.db.Prepare("insert into messages_for_delete (chat_id, message_id, add_date) values (?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(chatId, id, time.Now().Unix())
	if err != nil {
		return err
	}
	defer stmt.Close()
	return nil
}

func (DB *DB) deleteOldMessages(hours int64) ([][2]int64, error) {
	stmt, err := DB.db.Prepare("select id, chat_id, message_id from messages_for_delete where add_date <= ?")
	if err != nil {
		return nil, err
	}

	var id, chatId, msgId int64
	var ids []interface{}
	var idsMsg [][2]int64
	var bufArray [2]int64
	row, err := stmt.Query(time.Now().Unix() - 60*60*hours)
	for row.Next() {
		err := row.Scan(&id, &chatId, &msgId)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)

		bufArray[0] = chatId
		bufArray[1] = msgId
		idsMsg = append(idsMsg, bufArray)
	}

	if len(ids) != 0 {
		inStmtString := "in ("
		for i := 1; i <= len(ids); i++ {
			if i == len(ids) {
				inStmtString += "?)"
			} else {
				inStmtString += "?, "
			}
		}

		stmt, err = DB.db.Prepare("delete from main.messages_for_delete where id " + inStmtString)
		if err != nil {
			return nil, err
		}

		row, err = stmt.Query(ids...)
		if err != nil {
			return nil, err
		}
		defer row.Close()
		defer stmt.Close()
	}

	return idsMsg, nil
}
