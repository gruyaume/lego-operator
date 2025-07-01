package charm_test

import (
	"os"
	"testing"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/goopstest"
	"github.com/gruyaume/lego-operator/internal/charm"
)

func TestGivenNotLeaderWhenConfigureThenStatusBlocked(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)

	stateIn := goopstest.State{
		Leader: false,
	}

	stateOut := ctx.Run("start", stateIn)

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusBlocked,
		Message: "Unit is not leader",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenInvalidConfigWhenConfigureThenStatusBlocked(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"email":                   "invalid-email",
			"server":                  "",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
	}

	stateOut := ctx.Run("start", stateIn)

	if ctx.CharmErr != nil {
		t.Fatalf("unexpected charm error: %v", ctx.CharmErr)
	}

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusBlocked,
		Message: "Invalid config options: server config is empty",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenConfigureThenStatusActive(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"email":                   "guillaume@pizza.com",
			"server":                  "https://example.com",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
		Secrets: []goopstest.Secret{
			{
				ID: "some-secret-id",
				Content: map[string]string{
					"AWS_ACCESS_KEY_ID":   "AKIAIOSFODNN7EXAMPLE",
					"AWS_ASSUME_ROLE_ARN": "arn:aws:iam::123456789012:role/ExampleRole",
				},
			},
		},
	}

	stateOut := ctx.Run("start", stateIn)

	if ctx.CharmErr != nil {
		t.Fatalf("unexpected charm error: %v", ctx.CharmErr)
	}

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusActive,
		Message: "Certificates synchronized successfully",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected status %s, got %s", goops.StatusActive, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenConfigureThenEnvironmentVariablesAreSet(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"email":                   "guillaume@pizza.com",
			"server":                  "https://example.com",
			"plugin":                  "some-plugin",
			"plugin-config-secret-id": "some-secret-id",
		},
		Secrets: []goopstest.Secret{
			{
				ID: "some-secret-id",
				Content: map[string]string{
					"AWS_ACCESS_KEY_ID":   "AKIAIOSFODNN7EXAMPLE",
					"AWS_ASSUME_ROLE_ARN": "arn:aws:iam::123456789012:role/ExampleRole",
				},
			},
		},
	}

	_ = ctx.Run("start", stateIn)

	if ctx.CharmErr != nil {
		t.Fatalf("unexpected charm error: %v", ctx.CharmErr)
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
