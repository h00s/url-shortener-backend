package link

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/h00s/url-shortener-backend/db"
	"github.com/h00s/url-shortener-backend/host"
)

// Link represent one shortened link
type Link struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url" binding:"required"`
	password      string
	clientAddress string
	CreatedAt     string `json:"createdAt"`
}

func getLinkByID(db *db.Database, id int) (*Link, error) {
	return getLink(db, sqlGetLinkByID, fmt.Sprint(id))
}

func getLinkByName(db *db.Database, name string) (*Link, error) {
	return getLink(db, sqlGetLinkByName, strings.TrimSpace(name))
}

func getLinkByURL(db *db.Database, url string) (*Link, error) {
	return getLink(db, sqlGetLinkByURL, strings.TrimSpace(url))
}

func getLink(db *db.Database, query string, param string) (*Link, error) {
	l := &Link{}

	err := db.Conn.QueryRow(query, param).Scan(&l.ID, &l.Name, &l.URL, &l.password, &l.clientAddress, &l.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.New("Error while getting link (" + err.Error() + ")")
	}
	return l, nil
}

// InsertLink in db. If inserted, return Link struct
func insertLink(db *db.Database, url string, password string, clientAddress string) (*Link, error) {
	url = strings.TrimSpace(url)

	err := host.IsValid(url)
	if err != nil {
		return nil, errors.New("Link is invalid: " + err.Error())
	}

	// Check if URL is already in DB
	l, err := getLinkByURL(db, url)
	switch {
	case err != nil:
		return nil, err
	case l != nil:
		return l, nil
	}

	// URL is not in DB, insert it
	l = &Link{}
	id := 0

	password = strings.TrimSpace(password)
	hashedPassword := []byte("")
	if password != "" {
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("Error while creating password hash: " + err.Error())
		}
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(sqlInsertLink, nil, url, hashedPassword, clientAddress, "NOW()").Scan(&id)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("Error while inserting link: " + err.Error())
	}

	_, err = tx.Exec(sqlUpdateLinkName, getNameFromID(id), id)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("Error while updating link name: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	l, err = getLinkByID(db, id)
	if err != nil {
		return nil, errors.New("Error while getting link:" + err.Error())
	}

	return l, nil
}

// getNameFromID gets name from numerical ID
func getNameFromID(id int) string {
	name := ""
	for id > 0 {
		name = string(validChars[id%len(validChars)]) + name
		id = id / len(validChars)
	}
	return name
}

// getIDFromName gets ID from name
func getIDFromName(name string) int {
	id := 0
	for i := 0; i < len(name); i++ {
		id = len(validChars)*id + (strings.Index(validChars, string(name[i])))
	}
	return id
}
