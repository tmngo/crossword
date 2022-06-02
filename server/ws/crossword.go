package ws

type Puzzle struct {
	ID          string `json:"id"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	NumClues    int    `json:"numClues"`
	Grid        string `json:"grid"`
	AcrossClues []Clue `json:"acrossClues"`
	DownClues   []Clue `json:"downClues"`
	Title       string `json:"title"`
	Creators    string `json:"creator"`
	Attribution string `json:"attribution"`
	Gext        string `json:"gext"`
}

type PuzzleData struct {
	ID         string  `json:"id"`
	Source     string  `json:"source"`
	Year       int     `json:"year"`
	Month      int     `json:"month"`
	Day        int     `json:"day"`
	Title      string  `json:"title"`
	State      string  `json:"state"`
	Completion float64 `json:"completion"`
}

type Clue struct {
	Number    int       `json:"number"`
	Text      string    `json:"text"`
	Direction Direction `json:"direction"`
	Row       int       `json:"row"`
	Column    int       `json:"column"`
	Length    int       `json:"length"`
}

type Color struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

type Direction int

const (
	Across Direction = iota
	Down
)

func (d Direction) flip() Direction {
	return 1 - d
}

type Player struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Color    Color    `json:"color"`
	Position Position `json:"position"`
}

type Position struct {
	Row int       `json:"row"`
	Col int       `json:"col"`
	Dir Direction `json:"dir"`
}
