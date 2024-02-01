package utils

 import (
 		"time"
		"net"
    "net/http"
		"strings"

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
	if data.UserIP == "" || data.Location.Lat == "" || data.Location.Long == ""  || data.Title == "" || data.Text == "" {
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
