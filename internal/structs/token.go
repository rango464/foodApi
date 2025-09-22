package structs

type (
	AuthTokens struct {
		AccessToken  string `json:"Access"`
		RefreshToken string `json:"Refresh"`
	}
	TokenClaims struct {
		Sub  uint64 // ид пользователя
		Iat  uint64 // время создания токена
		Root uint64 // доступ (0-superadmin 1-admin 3-moderator 5-superuser 7-simpleuser 9-guest)
	}
)
