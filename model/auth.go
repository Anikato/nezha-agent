package model

import (
	"context"
	"errors"
	"strings"
)

type AuthHandler struct {
	clientSecret             string
	clientUUID               string
	xPanelName               string
	nodeRole                 string
	requireTransportSecurity bool
}

var ErrAuthCredentialsNotConfigured = errors.New("agent: AuthHandler credentials are not configured")

func normalizeNodeRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "xpanel", "openwrt":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func NewAuthHandler(clientSecret, clientUUID string, requireTransportSecurity bool, xPanelName, nodeRole string) *AuthHandler {
	return &AuthHandler{
		clientSecret:             clientSecret,
		clientUUID:               clientUUID,
		xPanelName:               strings.TrimSpace(xPanelName),
		nodeRole:                 normalizeNodeRole(nodeRole),
		requireTransportSecurity: requireTransportSecurity,
	}
}

func (a *AuthHandler) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if a == nil || a.clientSecret == "" || a.clientUUID == "" {
		return nil, ErrAuthCredentialsNotConfigured
	}
	metadata := map[string]string{
		"client-secret": a.clientSecret,
		"client-uuid":   a.clientUUID,
		"client_secret": a.clientSecret,
		"client_uuid":   a.clientUUID,
	}
	if a.xPanelName != "" {
		metadata["xpanel-name"] = a.xPanelName
		metadata["xpanel_name"] = a.xPanelName
	}
	if a.nodeRole != "" {
		metadata["node-role"] = a.nodeRole
		metadata["node_role"] = a.nodeRole
	}
	return metadata, nil
}

func (a *AuthHandler) RequireTransportSecurity() bool {
	if a == nil {
		return false
	}
	return a.requireTransportSecurity
}
