package domain

type Item struct {
	ID         int
	Name       string
	CategoryID int
}

type Category struct {
	ID   int
	Name string
}
