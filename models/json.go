package model

type Parameters struct {
	Body string `json:"body"`
}

type JsonResponse struct {
	CleanBody string `json:"cleaned_body,omitempty"`
	Error     string `json:"error,omitempty"`
}
