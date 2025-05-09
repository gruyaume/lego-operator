package charm

import (
	"context"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"go.opentelemetry.io/otel"
)

func HandleDefaultHook(ctx context.Context, hookContext *goops.HookContext) {
	return
}

func SetStatus(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("lego").Start(ctx, "SetStatus")
	defer span.End()

	status := commands.StatusActive

	message := ""

	statusSetOpts := &commands.StatusSetOptions{
		Name:    status,
		Message: message,
	}

	err := hookContext.Commands.StatusSet(statusSetOpts)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not set status:", err.Error())
		return
	}

	hookContext.Commands.JujuLog(commands.Info, "Status set to active")
}
