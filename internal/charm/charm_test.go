package charm_test

import (
	"os"
	"testing"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/goopstest"
	"github.com/gruyaume/lego-operator/internal/charm"
)

func TestGivenNotLeaderWhenConfigureThenStatusBlocked(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: false,
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateOut.UnitStatus != string(goops.StatusBlocked) {
		t.Errorf("expected status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenInvalidConfigWhenConfigureThenStatusBlocked(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
			"email":                   "invalid-email",
			"server":                  "",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateOut.UnitStatus != string(goops.StatusBlocked) {
		t.Errorf("expected status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenConfigureThenStatusActive(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
			"email":                   "guillaume@pizza.com",
			"server":                  "https://example.com",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
		Secrets: []*goopstest.Secret{
			{
				ID: "some-secret-id",
				Content: map[string]string{
					"AWS_ACCESS_KEY_ID":   "AKIAIOSFODNN7EXAMPLE",
					"AWS_ASSUME_ROLE_ARN": "arn:aws:iam::123456789012:role/ExampleRole",
				},
			},
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.CharmErr != nil {
		t.Fatalf("unexpected charm error: %v", ctx.CharmErr)
	}

	if stateOut.UnitStatus != string(goops.StatusActive) {
		t.Errorf("expected status %s, got %s", goops.StatusActive, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenConfigureThenEnvironmentVariablesAreSet(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
			"email":                   "guillaume@pizza.com",
			"server":                  "https://example.com",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
		Secrets: []*goopstest.Secret{
			{
				ID: "some-secret-id",
				Content: map[string]string{
					"AWS_ACCESS_KEY_ID":   "AKIAIOSFODNN7EXAMPLE",
					"AWS_ASSUME_ROLE_ARN": "arn:aws:iam::123456789012:role/ExampleRole",
				},
			},
		},
	}

	_, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envAWSKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if envAWSKey != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("expected AWS_ACCESS_KEY_ID to be 'AKIAIOSFODNN7EXAMPLE', got '%s'", envAWSKey)
	}

	envAWSRole := os.Getenv("AWS_ASSUME_ROLE_ARN")
	if envAWSRole != "arn:aws:iam::123456789012:role/ExampleRole" {
		t.Errorf("expected AWS_ASSUME_ROLE_ARN to be 'arn:aws:iam::123456789012:role/ExampleRole', got '%s'", envAWSRole)
	}
}
