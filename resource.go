package miri

type Resource interface {
	GetTitle() string
	GetType() string
	GetSongs() []*Song
	SetSongs(songs []*Song)
	Unmarshal(data []byte) error
}
