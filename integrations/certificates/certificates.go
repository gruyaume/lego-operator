package certificates

import (
	"encoding/json"
	"fmt"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
)

type IntegrationProvider struct {
	HookContext  *goops.HookContext
	RelationName string
}

type RequirerCertificateSigningRequests struct {
	CA                        bool   `json:"ca"`
	CertificateSigningRequest string `json:"certificate_signing_request"`
}

func (i *IntegrationProvider) GetRelationID() (string, error) {
	relationIDs, err := i.HookContext.Commands.RelationIDs(&commands.RelationIDsOptions{
		Name: i.RelationName,
	})
	if err != nil {
		return "", fmt.Errorf("could not get relation IDs: %w", err)
	}

	if len(relationIDs) == 0 {
		return "", fmt.Errorf("no relation IDs found for %s", i.RelationName)
	}

	return relationIDs[0], nil
}

func (i *IntegrationProvider) GetCertificateRequests() ([]*RequirerCertificateSigningRequests, error) {
	relationID, err := i.GetRelationID()
	if err != nil {
		return nil, fmt.Errorf("could not get relationID: %w", err)
	}

	relations, err := i.HookContext.Commands.RelationList(&commands.RelationListOptions{
		ID: relationID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not list relations for ID %s: %v", relationID, err)
	}

	if len(relations) == 0 {
		return nil, fmt.Errorf("no relations found for ID %s", relationID)
	}

	relationData, err := i.HookContext.Commands.RelationGet(&commands.RelationGetOptions{
		ID:     relationID,
		UnitID: relations[0],
		App:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get relation data: %w", err)
	}

	requestsStr := relationData["certificate_signing_requests"]
	if requestsStr == "" {
		return nil, fmt.Errorf("no request found in relation data")
	}

	var requirerCertificateSigningRequests []*RequirerCertificateSigningRequests

	err = json.Unmarshal([]byte(requestsStr), &requirerCertificateSigningRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal requirer requests: %w", err)
	}

	return requirerCertificateSigningRequests, nil
}

type ProviderAppData struct {
	CA                        string   `json:"ca"`
	Chain                     []string `json:"chain"`
	CertificateSigningRequest string   `json:"certificate_signing_request"`
	Certificate               string   `json:"certificate"`
}

type Certificate struct {
	Raw                 string
	CommonName          string
	ExpiryTime          string
	ValidityStartTime   string
	IsCA                bool
	SansDNS             []string
	SansIP              []string
	SansOID             []string
	EmailAddress        string
	Organization        string
	OrganizationalUnit  string
	CountryName         string
	StateOrProvinceName string
	LocalityName        string
}

type CertificateSigningRequest struct {
	Raw                 string
	CommonName          string
	SansDNS             []string
	SansIP              []string
	SansOID             []string
	EmailAddress        string
	Organization        string
	OrganizationalUnit  string
	CountryName         string
	StateOrProvinceName string
	LocalityName        string
}

type ProviderCertificate struct {
	RelationID                string
	Certificate               Certificate
	CertificateSigningRequest CertificateSigningRequest
	CA                        Certificate
	Chain                     []Certificate
	Revoked                   bool
}

func (i *IntegrationProvider) SetRelationCertificate(providerCertificate *ProviderCertificate) error {
	appData := []ProviderAppData{
		{
			CA:                        providerCertificate.CA.Raw,
			Chain:                     []string{},
			CertificateSigningRequest: providerCertificate.CertificateSigningRequest.Raw,
			Certificate:               providerCertificate.Certificate.Raw,
		},
	}
	for _, cert := range providerCertificate.Chain {
		appData[0].Chain = append(appData[0].Chain, cert.Raw)
	}

	appDataJSON, err := json.Marshal(appData)
	if err != nil {
		return fmt.Errorf("could not marshal app data: %w", err)
	}

	relationData := map[string]string{
		"certificates": string(appDataJSON),
	}

	relationSetOpts := &commands.RelationSetOptions{
		ID:   providerCertificate.RelationID,
		App:  true,
		Data: relationData,
	}

	err = i.HookContext.Commands.RelationSet(relationSetOpts)
	if err != nil {
		return fmt.Errorf("could not set relation data: %w", err)
	}

	return nil
}
