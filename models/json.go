package model

// type Parameters struct {
// 	Body string `json:"body"`
// 	Id   string `json:"user_id"`
// }

type JsonResponse struct {
	CleanBody string `json:"cleaned_body"`
}

type JsonErrorResponse struct {
	Error string `json:"error"`
}
