package db

import (
	"testing"

	"github.com/h00s/url-shortener-backend/config"
)

func TestDB(t *testing.T) {
	c, err := config.Load("../configuration_test.toml")
	if err != nil {
		t.Error("Unable to load configuration")
	}

	db, err := Connect(c.Database)
	if err != nil {
		t.Error("Unable to connect to DB", err)
	}

	db.Conn.Query("DROP TABLE schema; DROP TABLE links; DROP TABLE activities;")

	err = db.Migrate()
	if err != nil {
		t.Error("Unable to migrate DB", err)
	}
}
