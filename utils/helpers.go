package util

import "strings"

func CleanBody(body string) string {
	words := strings.Split(body, " ")
	i := 0
	var word string
	for i < len(words) {
		word = strings.ToLower(words[i])
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
		i++
	}
	return strings.Join(words, " ")
}
