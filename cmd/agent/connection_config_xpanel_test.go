package main

import (
	"context"
	"testing"

	"github.com/nezhahq/agent/model"
)

func TestConnectionConfigCarriesXPanelNameMetadata(t *testing.T) {
	restoreConnectionGenerationGlobals(t)
	publishRuntimeConfig(model.AgentConfig{
		Server:       "dashboard.example:443",
		ClientSecret: "secret",
		UUID:         "uuid",
		TLS:          true,
		XPanelName:   "私人面板",
	})

	metadata, err := loadConnectionConfigTuple().Auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata["xpanel-name"] != "私人面板" || metadata["xpanel_name"] != "私人面板" {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestConnectionConfigCarriesNodeRoleMetadata(t *testing.T) {
	restoreConnectionGenerationGlobals(t)
	publishRuntimeConfig(model.AgentConfig{
		Server:       "dashboard.example:443",
		ClientSecret: "secret",
		UUID:         "uuid",
		TLS:          true,
		NodeRole:     "openwrt",
	})

	metadata, err := loadConnectionConfigTuple().Auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata["node-role"] != "openwrt" || metadata["node_role"] != "openwrt" {
		t.Fatalf("metadata = %#v", metadata)
	}
}
