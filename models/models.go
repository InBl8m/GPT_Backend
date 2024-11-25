package models

type Page struct {
	ID        int    `json:"id"`
	HTML      string `json:"html"`
	Processed bool   `json:"processed"`
}

type SubmitRequest struct {
	ID        int    `json:"id"`
	Request   string `json:"request"`
	Processed bool   `json:"processed"`
}

type Answer struct {
	ID        int    `json:"id"`
	Answer    string `json:"answer"`
	Processed bool   `json:"processed"`
}
