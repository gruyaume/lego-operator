package charm

import (
	"fmt"
	"os"
	"strings"

	"github.com/gruyaume/charm-libraries/certificates"
	"github.com/gruyaume/goops"
	"github.com/gruyaume/lego-operator/internal/lego"
)

const (
	CertificatesIntegration = "certificates"
)

func Configure() error {
	isLeader, err := goops.IsLeader()
	if err != nil {
		return fmt.Errorf("could not check if unit is leader: %w", err)
	}

	if !isLeader {
		_ = goops.SetUnitStatus(goops.StatusBlocked, "Unit is not leader")
		return nil
	}

	goops.LogDebugf("Unit is leader")

	config := &ConfigOptions{}

	config.LoadFromJuju()

	err = config.Validate()
	if err != nil {
		_ = goops.SetUnitStatus(goops.StatusBlocked, fmt.Sprintf("Invalid config options: %s", err.Error()))
		return nil
	}

	goops.LogDebugf("Config is valid")

	err = setEnvVars(config.pluginConfigSecretID)
	if err != nil {
		_ = goops.SetUnitStatus(goops.StatusBlocked, fmt.Sprintf("Could not set environment variables: %s", err.Error()))
		return nil
	}

	err = syncCertificates()
	if err != nil {
		return fmt.Errorf("could not synchronize certificates: %w", err)
	}

	_ = goops.SetUnitStatus(goops.StatusActive, "Certificates synchronized successfully")

	return nil
}

// setEnvVars sets environment variables from the plugin configuration secret.
// It retrieves the secret content and sets each key-value pair as an environment variable.
func setEnvVars(pluginConfigSecretID string) error {
	secretContent, err := goops.GetSecretByID(pluginConfigSecretID, false, true)
	if err != nil {
		return fmt.Errorf("could not get secret %s: %w", pluginConfigSecretID, err)
	}

	for key, value := range secretContent {
		envKey := toEnvVarName(key)

		err := os.Setenv(envKey, value)
		if err != nil {
			goops.LogErrorf("Could not set environment variable %s: %s", envKey, err.Error())
			continue
		}
	}

	goops.LogDebugf("Environment variables set successfully")

	return nil
}

func toEnvVarName(key string) string {
	key = os.ExpandEnv(key)
	key = os.Expand(key, func(s string) string {
		return os.Getenv(s)
	})
	key = strings.ToUpper(key)
	key = strings.ReplaceAll(key, "-", "_")

	return key
}

func syncCertificates() error {
	certsIntegration := certificates.IntegrationProvider{
		RelationName: CertificatesIntegration,
	}

	certRequests, err := certsIntegration.GetOutstandingCertificateRequests()
	if err != nil {
		return fmt.Errorf("could not get certificate requests: %w", err)
	}

	if len(certRequests) == 0 {
		goops.LogInfof("No certificate requests found")
		return nil
	}

	config := &ConfigOptions{}

	config.LoadFromJuju()

	for _, cert := range certRequests {
		legoResponse, err := lego.RequestCertificate(config.email, config.server, cert.CertificateSigningRequest.Raw, config.plugin)
		if err != nil {
			goops.LogErrorf("Could not request certificate to acme server: %v", err.Error())
			continue
		}

		err = certsIntegration.SetRelationCertificate(&certificates.SetRelationCertificateOptions{
			RelationID:                cert.RelationID,
			CA:                        legoResponse.IssuerCertificate,
			Chain:                     []string{legoResponse.IssuerCertificate},
			CertificateSigningRequest: cert.CertificateSigningRequest.Raw,
			Certificate:               legoResponse.Certificate,
		})
		if err != nil {
			goops.LogErrorf("Could not set certificate in relation data: %v", err.Error())
			continue
		}

		goops.LogInfof("Successfully set certificate for relation %s", cert.RelationID)
	}

	return nil
}
