package logging

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// getLogIP extracts and formats the client's IP address from an HTTP request.
// It checks the "X-Forwarded-For" header for the IP address if present, otherwise
// it uses the RemoteAddr field. The function removes the port number from the IP
// address and adjusts its length based on whether it contains a comma (indicating
// multiple IPs). The function returns two versions of the IP address: a colored
// version for logging purposes and a plain version.
//
// Parameters:
//   - Request: A pointer to an http.Request object containing the client's request.
//
// Returns:
//   - A string representing the colored IP address.
//   - A string representing the plain IP address.
func getLogIP(Request *http.Request) (string, string) {

	// Get the IP address of the client
	ip := Request.RemoteAddr
	if forwarded := Request.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded // If the X-Forwarded-For header is present, use it instead
	}

	// Remove the port number from the IP address
	ip = regexp.MustCompile(`:\d+$`).ReplaceAllString(ip, "")
	ip = "  " + ip

	// If the IP address contains a comma
	if strings.Contains(ip, ",") {
		ip = fixStringLength(ip, 34)
	} else {
		ip = fixStringLength(ip, 15)
	}

	colored := color.MagentaString(ip)
	// Return the colored IP address and the non-colored IP address
	return colored, ip
}
