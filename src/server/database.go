package server

import (
	"database/sql"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
)

type database struct {
	*sql.DB
}

func newDatabase(dataSourceName string) (*database, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(4 * time.Minute)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)

	// Create table to save subscribed channel data
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS channels (id VARCHAR(255) PRIMARY KEY, title TEXT);")
	if err != nil {
		return nil, err
	}

	// Create table to save subscribing chat data
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS chats (id BIGINT PRIMARY KEY, admin BOOL DEFAULT 0);")
	if err != nil {
		return nil, err
	}

	// Create table to save subscribers data pairs
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS subscribers (" +
		"chatID BIGINT, channelID VARCHAR(255), PRIMARY KEY (chatID, channelID));")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS monitoring (" +
		"videoID VARCHAR(255), chatID BIGINT, messageID INT, PRIMARY KEY (videoID, chatID));")
	if err != nil {
		return nil, err
	}

	// Create table to save videos status
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS videos (id VARCHAR(255) PRIMARY KEY, completed BOOL, title TEXT, startTime BIGINT, channelID TEXT);")
	if err != nil {
		return nil, err
	}

	return &database{DB: db}, nil
}

// Subscribe registers info into corresponding table
func (db *database) subscribe(chat rowChat, channel rowChannel) error {
	_, err := db.Exec("INSERT IGNORE INTO chats (id, admin) VALUES (?, ?);", chat.id, false)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT IGNORE INTO channels (id, title) VALUES (?, ?);", channel.id, channel.title)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT IGNORE INTO subscribers (chatID, channelID) VALUES (?, ?);", chat.id, channel.id)
	if err != nil {
		return err
	}

	return nil
}

func (db *database) getSubscribeChatsByChannelID(channelID string) ([]rowChat, error) {
	var results []rowChat

	if err := db.QueryResults(
		&results,
		func(rows *sql.Rows, dest interface{}) error {
			chat := dest.(*rowChat)
			return rows.Scan(&chat.id, &chat.admin)
		},
		"SELECT chats.id, chats.admin FROM "+
			"chats INNER JOIN subscribers ON chats.id = subscribers.chatID "+
			"WHERE subscribers.channelID = ?;",
		channelID,
	); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *database) getSubscribedChannels() ([]rowChannel, error) {
	var results []rowChannel

	if err := db.QueryResults(
		&results,
		func(rows *sql.Rows, dest interface{}) error {
			r := dest.(*rowChannel)
			return rows.Scan(&r.id, &r.title)
		},
		"SELECT id, title FROM channels;",
	); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *database) getSubscribedChannelsByChatID(chatID int64) ([]rowChannel, error) {
	var results []rowChannel

	if err := db.QueryResults(
		&results,
		func(rows *sql.Rows, dest interface{}) error {
			ch := dest.(*rowChannel)
			return rows.Scan(&ch.id, &ch.title)
		},
		"SELECT channels.id, channels.title FROM "+
			"channels INNER JOIN subscribers ON channels.id = subscribers.channelID "+
			"WHERE subscribers.chatID = ?;",
		chatID,
	); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *database) getMonitoringMessagesByVideoID(videoID string) ([]rowMonitoring, error) {
	var results []rowMonitoring

	if err := db.QueryResults(
		&results,
		func(rows *sql.Rows, dest interface{}) error {
			r := dest.(*rowMonitoring)
			return rows.Scan(&r.videoID, &r.chatID, &r.messageID)
		},
		"SELECT videoID, chatID, messageID FROM monitoring WHERE videoID = ?;",
		videoID,
	); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *database) QueryResults(container interface{}, scan func(rows *sql.Rows, dest interface{}) error, query string, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	results := reflect.ValueOf(container).Elem()
	element := reflect.New(results.Type().Elem())
	for rows.Next() {
		if err := scan(rows, element.Interface()); err != nil {
			return err
		}

		results.Set(reflect.Append(results, element.Elem()))
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}
