package utils

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"time"

	"pawmap/database"
	"pawmap/message"
)

func IsRateLimited(userIP string) bool {

	lastCreationTime, err := GetLastCreationTime(userIP)

	if err != nil {
		return false
	}

	currentTime := time.Now()

	return currentTime.Sub(lastCreationTime) < time.Minute
}

func GetLastCreationTime(userIP string) (time.Time, error) {
	lastCreationTime, err := database.GetLastCreatedTime(userIP)

	if err != nil {
		return time.Now(), err
	}

	return lastCreationTime, err
}

func CheckMessageData(data *message.Pin) bool {
	if data.Location.Lat == "" || data.Location.Long == "" ||
		data.Title == "" || data.Text == "" {
		return false // not Valid
	}
	return true // Valid
}

func CheckUserData(data *message.User) bool {
	if data.Username == "" || data.Password == "" || data.Email == "" {
		return false // not Valid
	}
	return true // Valid
}

func GetPinModifier(pin *message.Pin) {
	pin.UserIP = ""
}

func GetIPAddress(r *http.Request) string {

	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		ip = r.RemoteAddr
	}

	ipParts := strings.Split(ip, ",")
	firstIP := strings.TrimSpace(ipParts[0])

	host, _, err := net.SplitHostPort(firstIP)
	if err == nil {
		return host
	}

	return firstIP
}

func Base64EncodeString(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func Base64DecodeString(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func CreateToken(username string, current_time string, end_time string) string {
	return username + "_" + current_time + "_" + end_time
}

func ValidateToken(token string) (string, string, string) {
	parts := strings.Split(token, "_")
	return parts[0], parts[1], parts[2]
}
