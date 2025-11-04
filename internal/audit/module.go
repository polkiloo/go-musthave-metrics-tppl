package audit

import "go.uber.org/fx"

// Module wires audit dependencies for Fx.
var Module = fx.Module(
	"audit",
	fx.Provide(
		func() Clock { return systemClock },
		NewPublisher,
	),
)
