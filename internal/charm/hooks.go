package charm

import (
	"context"
	"fmt"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/lego-operator/integrations/certificates"
	"go.opentelemetry.io/otel"
)

func HandleDefaultHook(ctx context.Context) {
	_, span := otel.Tracer("lego").Start(ctx, "handle default hook")
	defer span.End()

	err := ensureLeader(ctx)
	if err != nil {
		return
	}

	err = validateConfigOptions(ctx)
	if err != nil {
		return
	}

	err = syncCertificates(ctx)
	if err != nil {
		return
	}
}

func SetStatus(ctx context.Context) {
	_, span := otel.Tracer("lego").Start(ctx, "set status")
	defer span.End()

	status := goops.StatusActive
	message := ""

	err := validateConfigOptions(ctx)
	if err != nil {
		status = goops.StatusBlocked
		message = fmt.Sprintf("invalid config: %s", err.Error())
	}

	err = goops.SetUnitStatus(status, message)
	if err != nil {
		goops.LogErrorf("could not set status: %v", err)
		return
	}

	goops.LogInfof("Status set to %s: %s", status, message)
}

func ensureLeader(ctx context.Context) error {
	_, span := otel.Tracer("lego").Start(ctx, "ensure leader")
	defer span.End()

	isLeader, err := goops.IsLeader()
	if err != nil {
		goops.LogErrorf("could not check if unit is leader: %v", err)
		return fmt.Errorf("could not check if unit is leader: %w", err)
	}

	if !isLeader {
		goops.LogWarningf("Unit is not leader")
		return fmt.Errorf("unit is not leader")
	}

	goops.LogInfof("Unit is leader")

	return nil
}

func validateConfigOptions(ctx context.Context) error {
	_, span := otel.Tracer("lego").Start(ctx, "validate config")
	defer span.End()

	config := &ConfigOptions{}

	err := config.LoadFromJuju()
	if err != nil {
		goops.LogWarningf("Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		goops.LogWarningf("Config is not valid: %s", err.Error())
		return fmt.Errorf("config is not valid: %w", err)
	}

	return nil
}

func syncCertificates(ctx context.Context) error {
	_, span := otel.Tracer("lego").Start(ctx, "sync certificates")
	defer span.End()

	certsIntegration := certificates.IntegrationProvider{
		RelationName: "certificates",
	}

	certRequests, err := certsIntegration.GetCertificateRequests()
	if err != nil {
		goops.LogErrorf("Error getting certificate requests: %v", err)
		return err
	}

	if len(certRequests) == 0 {
		goops.LogInfof("No certificate requests found")
		return nil
	}

	config := &ConfigOptions{}

	err = config.LoadFromJuju()
	if err != nil {
		goops.LogWarningf("Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	for _, cert := range certRequests {
		_, err := requestCertificate(config.email, config.server, cert.CertificateSigningRequest, config.plugin)
		if err != nil {
			goops.LogErrorf("Could not request certificate to acme server: %v", err.Error())
			continue
		}

		relationID, err := certsIntegration.GetRelationID()
		if err != nil {
			goops.LogErrorf("Could not get relation ID: %v", err.Error())
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
			goops.LogErrorf("Could not set certificate in relation data: %v", err.Error())
			continue
		}
	}

	return nil
}
