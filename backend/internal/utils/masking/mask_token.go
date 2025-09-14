package masking

// MaskToken masks the token by keeping only the last 4 characters visible.
// If the token is shorter than 4 characters, it masks all but the last character.
func MaskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) < 4 {
		return "***" + token[len(token)-1:]
	}
	return "***" + token[len(token)-4:]
}
