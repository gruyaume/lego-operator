package charm

import (
	"fmt"

	"github.com/gruyaume/charm-libraries/certificates"
	"github.com/gruyaume/goops"
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

	err = config.LoadFromJuju()
	if err != nil {
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		_ = goops.SetUnitStatus(goops.StatusBlocked, fmt.Sprintf("Invalid config options: %s", err.Error()))
		return nil
	}

	err = syncCertificates()
	if err != nil {
		return fmt.Errorf("could not synchronize certificates: %w", err)
	}

	_ = goops.SetUnitStatus(goops.StatusActive, "Certificates synchronized successfully")

	return nil
}

func syncCertificates() error {
	certsIntegration := certificates.IntegrationProvider{
		RelationName: "certificates",
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

	err = config.LoadFromJuju()
	if err != nil {
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	for _, cert := range certRequests {
		legoResponse, err := requestCertificate(config.email, config.server, cert.CertificateSigningRequest.Raw, config.plugin)
		if err != nil {
			goops.LogErrorf("Could not request certificate to acme server: %v", err.Error())
			continue
		}

		err = certsIntegration.SetRelationCertificate(&certificates.SetRelationCertificateOptions{
			RelationID:                cert.RelationID,
			CA:                        "",         // TODO: Fill in properly
			Chain:                     []string{}, // TODO: Fill in properly
			CertificateSigningRequest: cert.CertificateSigningRequest.Raw,
			Certificate:               legoResponse.Certificate,
		})
		if err != nil {
			goops.LogErrorf("Could not set certificate in relation data: %v", err.Error())
			continue
		}
	}

	return nil
}
