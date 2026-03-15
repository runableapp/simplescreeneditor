package editor

type Cursor struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func (c *Cursor) Clamp(rows, cols int) {
	if c.Row < 0 {
		c.Row = 0
	}
	if c.Col < 0 {
		c.Col = 0
	}
	if c.Row >= rows {
		c.Row = rows - 1
	}
	if c.Col >= cols {
		c.Col = cols - 1
	}
}
