package alpsviewmultipart

import (
	"git.sr.ht/~migadu/alps"
)

func init() {
	p := alps.GoPlugin{Name: "viewmultipart"}
	alps.RegisterPluginLoader(p.Loader())
}
