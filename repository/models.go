package repository

// Game represents the Game table in database
type Game struct {
	ID           string `json:"id,omitempty"`
	Board        string `json:"board,omitempty"`
	Status       string `json:"status,omitempty"`
	ComputerMark string `json:"-"`
}
