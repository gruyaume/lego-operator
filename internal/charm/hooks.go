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

	err := ensureLeader(ctx, hookContext)
	if err != nil {
		return
	}

	err = validateConfigOptions(ctx, hookContext)
	if err != nil {
		return
	}

	err = syncCertificates(ctx, hookContext)
	if err != nil {
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

func ensureLeader(ctx context.Context, hookContext *goops.HookContext) error {
	_, span := otel.Tracer("lego").Start(ctx, "ensure leader")
	defer span.End()

	isLeader, err := hookContext.Commands.IsLeader()
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Could not check if unit is leader:", err.Error())
		return fmt.Errorf("could not check if unit is leader: %w", err)
	}

	if !isLeader {
		hookContext.Commands.JujuLog(commands.Warning, "Unit is not leader")
		return fmt.Errorf("unit is not leader")
	}

	hookContext.Commands.JujuLog(commands.Info, "Unit is leader")

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

	for _, cert := range certRequests {
		_, err := requestCertificate("", "", cert.CertificateSigningRequest, "") // TO DO
		if err != nil {
			hookContext.Commands.JujuLog(commands.Error, "Could not request certificate: %v", err.Error())
			continue
		}
	}

	return nil
}
