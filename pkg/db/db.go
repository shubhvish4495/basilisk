package db

import "context"

type DB interface {
	Ping() error
	PingContext(ctx context.Context) error
	Close() error
	Init() error
}

var instance DB

func GetDb() (DB, func() error) {
	if instance == nil {
		return nil, nil
	}
	return instance, instance.Close
}

func SetDB(db DB) {
	instance = db
}
