package lego

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
)

type LetsEncryptUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

type LegoOutputResponse struct {
	CSR               string `json:"csr"`
	PrivateKey        string `json:"private_key"`
	Certificate       string `json:"certificate"`
	IssuerCertificate string `json:"issuer_certificate"`
	Metadata          `json:"metadata"`
}

type Metadata struct {
	StableURL string `json:"stable_url"`
	URL       string `json:"url"`
	Domain    string `json:"domain"`
}

func (u *LetsEncryptUser) GetEmail() string {
	return u.Email
}

func (u LetsEncryptUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *LetsEncryptUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func RequestCertificate(email, server, csr, plugin string) (*LegoOutputResponse, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate priv key: %s", err)
	}

	user := LetsEncryptUser{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(&user)

	config.CADirURL = server
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("couldn't create lego client: %s", err)
	}

	dnsProvider, err := dns.NewDNSChallengeProviderByName(plugin)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("couldn't create %s provider: ", plugin), err)
	}

	err = client.Challenge.SetDNS01Provider(dnsProvider)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("couldn't set %s DNS provider server: ", plugin), err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("couldn't register user: %s", err)
	}

	user.Registration = reg

	block, _ := pem.Decode([]byte(csr))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, errors.New("failed to decode PEM block containing certificate request")
	}

	csrObject, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %s", err)
	}

	certificates, err := client.Certificate.ObtainForCSR(certificate.ObtainForCSRRequest{
		CSR:    csrObject,
		Bundle: true,
	})
	if err != nil {
		return nil, fmt.Errorf("coudn't obtain cert: %s", err)
	}

	return &LegoOutputResponse{
		CSR:               string(certificates.CSR),
		PrivateKey:        string(certificates.PrivateKey),
		Certificate:       string(certificates.Certificate),
		IssuerCertificate: string(certificates.IssuerCertificate),
		Metadata: Metadata{
			StableURL: certificates.CertStableURL,
			URL:       certificates.CertURL,
			Domain:    certificates.Domain,
		},
	}, nil
}
