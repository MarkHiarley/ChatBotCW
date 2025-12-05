package models

type Document struct {
	ID      string
	Content string
	Source  string
	Vector  []float32
}
