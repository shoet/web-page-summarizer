package entities

type LoginSession struct {
	IdToken      string `json:"idToken"`
	AccessToken  string `jsonn:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
