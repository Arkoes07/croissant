package secret

// Secret is a domain that represents the structure of data in the secret file
type Secret struct {
	Spotify Spotify
}

// Spotify represents the structure of spotify secret data in the file
type Spotify struct {
	ClientID     string
	ClientSecret string
}

// Service is a port that defines available behavior of secret package
type Service interface {
	// Parse will parse and return the data in the secret file
	Parse() (*Secret, error)
}
