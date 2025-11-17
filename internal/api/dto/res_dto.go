package dto

type Res struct {
	Success bool `json:"success"`
	Error   any  `json:"error"`
	Data    any  `json:"data"`
}
