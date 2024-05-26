package utils

import (
	"pawmap/database"
	"pawmap/message"
	"strconv"
	"time"
)

func CreateSession(userID uint) string {
	var session message.Session

	session.UserID = userID
	session.StartedAt = time.Now()
	session.ExpiresAt = time.Now().Add(time.Hour * 2)

	if database.CreateSession(&session) == 1 {
		// Create Token For Session
		user, error := database.GetUserByID(session.UserID)
		if error != nil {
			return ""
		}

		token := Base64EncodeString(CreateToken(user.Username, strconv.FormatInt(session.StartedAt.Unix(), 10), strconv.FormatInt(session.ExpiresAt.Unix(), 10)))
		return token
	} else {
		return ""
	}
}
