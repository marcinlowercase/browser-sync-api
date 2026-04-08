package auth

type RequestCodePayload struct {
	Email string `json:"email"`
}

type VerifyCodePayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AuthResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
}
