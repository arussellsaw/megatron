package domain

import "time"

type Episode struct {
	ID        string
	Path      string
	Subtitles []Caption
}

type Caption struct {
	ID    string
	Start time.Duration
	End   time.Duration
	Text  []string
}
