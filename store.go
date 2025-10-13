package main

type machine struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	MAC     string `json:"mac"`
	IP      string `json:"ip"`
	Control string `json:"control"`
}
