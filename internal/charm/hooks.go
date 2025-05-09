package charm

import (
	"context"
	"fmt"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"github.com/gruyaume/lego-operator/integrations/certificates"
	"go.opentelemetry.io/otel"
)

func HandleDefaultHook(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("lego").Start(ctx, "handle default hook")
	defer span.End()

	// Validate config options
	err := validateConfigOptions(ctx, hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Invalid config options:", err.Error())
		return
	}

	err = syncCertificates(ctx, hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Error syncing certificates:", err.Error())
		return
	}
}

func SetStatus(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("lego").Start(ctx, "set status")
	defer span.End()

	status := commands.StatusActive
	message := ""

	err := validateConfigOptions(ctx, hookContext)
	if err != nil {
		status = commands.StatusBlocked
		message = fmt.Sprintf("invalid config: %s", err.Error())
	}

	err = hookContext.Commands.StatusSet(&commands.StatusSetOptions{
		Name:    status,
		Message: message,
	})
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not set status:", err.Error())
		return
	}

	hookContext.Commands.JujuLog(commands.Info, "Status set to active")
}

func syncCertificates(ctx context.Context, hookContext *goops.HookContext) error {
	_, span := otel.Tracer("lego").Start(ctx, "sync certificates")
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

func validateConfigOptions(ctx context.Context, hookContext *goops.HookContext) error {
	_, span := otel.Tracer("lego").Start(ctx, "validate config")
	defer span.End()

	email, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "email"})
	if err != nil {
		return fmt.Errorf("failed to get email config: %w", err)
	}

	if email == "" {
		return fmt.Errorf("email config is empty")
	}

	server, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "server"})
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	if server == "" {
		return fmt.Errorf("server config is empty")
	}

	return nil
}
