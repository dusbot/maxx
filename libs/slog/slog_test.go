package slog

import (
	"testing"

	"github.com/dusbot/maxx/libs/color"
)

func TestLevel(t *testing.T) {
	color.Enabled()
	Println(DEBUG, "test1")
	Println(INFO, "test1")
	Println(WARN, "test1")
	Println(ERROR, "test1")
	Println(DATA, "test1")
	Println(NONE, "test1")
}
