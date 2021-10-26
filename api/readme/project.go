package readme

type Project struct {
	Name      string `json:"name"`
	SubDomain string `json:"subdomain"`
	JWTSecret string `json:"jwtSecret"`
	BaseUrl   string `json:"baseUrl"`
	Plan      string `json:"plan"`
}
