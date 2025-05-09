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

	config := &ConfigOptions{}

	err := config.LoadFromJuju(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Config is not valid: %s", err.Error())
		return fmt.Errorf("config is not valid: %w", err)
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

	config := &ConfigOptions{}

	err = config.LoadFromJuju(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	for _, cert := range certRequests {
		_, err := requestCertificate(config.email, config.server, cert.CertificateSigningRequest, config.plugin)
		if err != nil {
			hookContext.Commands.JujuLog(commands.Error, "Could not request certificate to acme server: %v", err.Error())
			continue
		}

		relationID, err := certsIntegration.GetRelationID()
		if err != nil {
			hookContext.Commands.JujuLog(commands.Error, "Could not get relation ID: %v", err.Error())
			continue
		}

		err = certsIntegration.SetRelationCertificate(&certificates.ProviderCertificate{ // TODO: Fill in properly
			RelationID:                relationID,
			Certificate:               certificates.Certificate{},
			CertificateSigningRequest: certificates.CertificateSigningRequest{},
			CA:                        certificates.Certificate{},
			Chain:                     []certificates.Certificate{},
			Revoked:                   false,
		})
		if err != nil {
			hookContext.Commands.JujuLog(commands.Error, "Could not set certificate in relation data: %v", err.Error())
			continue
		}
	}

	return nil
}
