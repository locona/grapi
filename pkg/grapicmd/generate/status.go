package generate

import "github.com/izumin5210/grapi/pkg/grapicmd/ui"

type status int

const (
	statusCreate status = iota
	statusExist
	statusIdentical
	statusConflicted
	statusForce
	statusSkipped
)

var (
	creatableStatusSet = map[status]struct{}{
		statusCreate: {},
		statusForce:  {},
	}
)

func (s status) Fprint(ui ui.UI, msg string) {
	switch s {
	case statusCreate, statusForce:
		ui.ItemSuccess(msg)
	case statusConflicted:
		ui.ItemFailure(msg)
	default:
		ui.ItemSkipped(msg)
	}
}

func (s status) ShouldCreate() bool {
	_, ok := creatableStatusSet[s]
	return ok
}
