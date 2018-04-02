package link

import (
	"github.com/gin-gonic/gin"
	"github.com/h00s/url-shortener-backend/db"
	"github.com/h00s/url-shortener-backend/logger"
)

// Controller for Link methods
type Controller struct {
	db     *db.Database
	logger *logger.Logger
}

type errorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
}

// InsertLinkData represents data which is sent when posting new url
type InsertLinkData struct {
	URL string `json:"url" binding:"required"`
}

// NewController creates new link controller
func NewController(db *db.Database, logger *logger.Logger) *Controller {
	return &Controller{db: db, logger: logger}
}

// GetLink get link with specific name
func (lc *Controller) GetLink(c *gin.Context) {
	name := c.Param("name")
	l, err := getLinkByName(lc.db, name)

	switch {
	case l != nil:
		insertActivity(lc.db, l.ID, c.ClientIP())
		c.JSON(200, l)
	case err != nil:
		lc.logger.Error(err.Error())
		c.JSON(500, errorResponse{"Error while getting link", "There was an server error when getting link"})
	default:
		c.JSON(404, errorResponse{"Link not found", "Link with specified name not found "})
	}
}

// GetLinkActivityStats get link with specific name
func (lc *Controller) GetLinkActivityStats(c *gin.Context) {
	name := c.Param("name")
	id := getIDFromName(name)
	s, err := getLinkActivityStats(lc.db, id)

	switch {
	case s != nil:
		c.JSON(200, s)
	case err != nil:
		lc.logger.Error(err.Error())
		c.JSON(500, errorResponse{"Error while getting link", "There was an server error when getting link"})
	default:
		c.JSON(404, errorResponse{"Link not found", "Link with specified name not found "})
	}
}

// RedirectToLink redirects to link with specific name
func (lc *Controller) RedirectToLink(c *gin.Context) {
	name := c.Param("name")
	l, err := getLinkByName(lc.db, name)

	switch {
	case l != nil:
		insertActivity(lc.db, l.ID, c.ClientIP())
		c.Redirect(302, l.URL)
	case err != nil:
		lc.logger.Error(err.Error())
		c.JSON(500, errorResponse{"Error while getting link", "There was an server error when getting link"})
	default:
		c.JSON(404, errorResponse{"Link not found", "Link with specified name not found"})
	}
}

// InsertLink adds new link
func (lc *Controller) InsertLink(c *gin.Context) {
	if !lc.isSpammer(c.ClientIP()) {
		var linkData InsertLinkData
		if err := c.BindJSON(&linkData); err == nil {
			l, err := insertLink(lc.db, linkData.URL, c.ClientIP())
			if err == nil {
				c.JSON(201, l)
			} else {
				lc.logger.Error(err.Error())
				c.JSON(500, errorResponse{"Error while adding link", "There was an server error when adding link"})
			}
		} else {
			c.JSON(400, errorResponse{"Request is invalid", "Request should be a JSON object containing url"})
		}
	} else {
		c.JSON(429, errorResponse{"Rate limiting", "Too many links posted, please wait couple of minutes"})
	}
}

func (lc *Controller) isSpammer(clientAddress string) bool {
	linkCount := 0
	lc.db.Conn.QueryRow(sqlGetPostCountInLastMinutes, clientAddress, 10).Scan(&linkCount)

	if linkCount >= 10 {
		return true
	}
	return false
}
