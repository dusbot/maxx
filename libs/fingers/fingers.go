package fingers

import "github.com/chainreactors/fingers"

var Engine *fingers.Engine

func init() {
	Engine, _ = fingers.NewEngine()
}
