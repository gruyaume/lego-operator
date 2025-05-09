package charm

import (
	"context"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"github.com/gruyaume/lego-operator/integrations/certificates"
	"go.opentelemetry.io/otel"
)

func HandleDefaultHook(ctx context.Context, hookContext *goops.HookContext) {
	err := syncCertificates(ctx, hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Error syncing certificates:", err.Error())
		return
	}
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

func syncCertificates(ctx context.Context, hookContext *goops.HookContext) error {
	_, span := otel.Tracer("lego").Start(ctx, "syncCertificates")
	defer span.End()

	certsIntegration := certificates.IntegrationProvider{
		HookContext:  hookContext,
		RelationName: "certificates",
	}

	certRequests, err := certsIntegration.GetCertificateRequests()
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Error getting certificate requests:", err.Error())
		return err
	}

	if len(certRequests) == 0 {
		hookContext.Commands.JujuLog(commands.Info, "No certificate requests found")
		return nil
	}

	return nil
}
