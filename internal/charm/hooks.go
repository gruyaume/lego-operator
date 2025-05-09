package charm

import (
	"context"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"go.opentelemetry.io/otel"
)

func HandleDefaultHook(ctx context.Context, hookContext *goops.HookContext) {
}

func SetStatus(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("lego").Start(ctx, "SetStatus")
	defer span.End()

	err := hookContext.Commands.StatusSet(&commands.StatusSetOptions{
		Name:    commands.StatusActive,
		Message: "",
	})
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not set status:", err.Error())
		return
	}

	hookContext.Commands.JujuLog(commands.Info, "Status set to active")
}
