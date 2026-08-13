package model

import (
	"context"
	"sync"
	"testing"
)

func TestAuthHandlerCapturesImmutableConnectionCredentials(t *testing.T) {
	// Given
	secret := "secret-a"
	uuid := "uuid-a"
	auth := NewAuthHandler(secret, uuid, false, "", "")

	// When
	secret = "secret-b"
	uuid = "uuid-b"
	metadata, err := auth.GetRequestMetadata(context.Background())

	// Then
	if err != nil {
		t.Fatalf("GetRequestMetadata: %v", err)
	}
	if metadata["client-secret"] != "secret-a" || metadata["client-uuid"] != "uuid-a" {
		t.Fatalf("connection-scoped credentials changed after construction: %v", metadata)
	}
	if auth.RequireTransportSecurity() {
		t.Fatal("connection-scoped TLS policy changed after construction")
	}
}

func TestAuthHandlerEmitsLegacyAndHyphenatedMetadata(t *testing.T) {
	// Given
	auth := NewAuthHandler("secret", "uuid", true, "", "")

	// When
	metadata, err := auth.GetRequestMetadata(context.Background())

	// Then
	if err != nil {
		t.Fatalf("GetRequestMetadata: %v", err)
	}
	for _, key := range []string{"client-secret", "client_secret"} {
		if metadata[key] != "secret" {
			t.Fatalf("%s = %q, want secret", key, metadata[key])
		}
	}
	for _, key := range []string{"client-uuid", "client_uuid"} {
		if metadata[key] != "uuid" {
			t.Fatalf("%s = %q, want uuid", key, metadata[key])
		}
	}
	if !auth.RequireTransportSecurity() {
		t.Fatal("TLS-enabled connection must require transport security")
	}
}

func TestAuthHandlerGetRequestMetadataNilSafe(t *testing.T) {
	// Given
	var nilHandler *AuthHandler
	zero := &AuthHandler{}

	// When / Then
	if _, err := nilHandler.GetRequestMetadata(context.Background()); err == nil {
		t.Fatal("nil receiver must return an error, not panic")
	}
	if _, err := zero.GetRequestMetadata(context.Background()); err == nil {
		t.Fatal("zero-value AuthHandler must return an error, not panic")
	}
}

func TestAuthHandlerIntranetPlaintextStillAuthenticates(t *testing.T) {
	// Given
	auth := NewAuthHandler("intranet-secret", "intranet-uuid", false, "", "")

	// When
	metadata, err := auth.GetRequestMetadata(context.Background())

	// Then
	if err != nil {
		t.Fatalf("intranet agent must produce credentials: %v", err)
	}
	if auth.RequireTransportSecurity() {
		t.Fatal("TLS:false intranet connection must allow plaintext transport")
	}
	if metadata["client_secret"] != "intranet-secret" || metadata["client_uuid"] != "intranet-uuid" {
		t.Fatalf("intranet connection must send its captured credentials: %v", metadata)
	}
}

func TestAuthHandlerConcurrentReadsRemainImmutable(t *testing.T) {
	// Given
	auth := NewAuthHandler("secret-a", "uuid-a", true, "", "")
	const readers = 16
	const rounds = 1000
	var waitGroup sync.WaitGroup
	waitGroup.Add(readers)

	// When
	for range readers {
		go func() {
			defer waitGroup.Done()
			for range rounds {
				metadata, err := auth.GetRequestMetadata(context.Background())
				if err != nil {
					t.Errorf("GetRequestMetadata: %v", err)
					return
				}
				if metadata["client-secret"] != "secret-a" || metadata["client-uuid"] != "uuid-a" || !auth.RequireTransportSecurity() {
					t.Errorf("immutable handler changed: metadata=%v requireTLS=%v", metadata, auth.RequireTransportSecurity())
					return
				}
			}
		}()
	}
	waitGroup.Wait()

	// Then
	metadata, err := auth.GetRequestMetadata(context.Background())
	if err != nil || metadata["client-secret"] != "secret-a" || metadata["client-uuid"] != "uuid-a" {
		t.Fatalf("final immutable metadata mismatch: metadata=%v err=%v", metadata, err)
	}
}

func TestAuthHandlerRequireTransportSecurityNilSafe(t *testing.T) {
	// Given
	var nilHandler *AuthHandler
	zero := &AuthHandler{}

	// When / Then
	if nilHandler.RequireTransportSecurity() {
		t.Fatal("nil receiver must not require TLS")
	}
	if zero.RequireTransportSecurity() {
		t.Fatal("zero-value handler must preserve plaintext-compatible behavior")
	}
}

func TestAuthHandlerEmitsTrimmedXPanelNameAliases(t *testing.T) {
	auth := NewAuthHandler("secret", "uuid", true, " 私人面板 ", "")
	metadata, err := auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata["xpanel-name"] != "私人面板" || metadata["xpanel_name"] != "私人面板" {
		t.Fatalf("xpanel metadata = %#v", metadata)
	}
}

func TestAuthHandlerOmitsEmptyXPanelName(t *testing.T) {
	auth := NewAuthHandler("secret", "uuid", true, "  ", "")
	metadata, err := auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := metadata["xpanel-name"]; ok {
		t.Fatalf("unexpected xpanel-name metadata: %#v", metadata)
	}
	if _, ok := metadata["xpanel_name"]; ok {
		t.Fatalf("unexpected xpanel_name metadata: %#v", metadata)
	}
}

func TestAuthHandlerEmitsNodeRoleAliases(t *testing.T) {
	auth := NewAuthHandler("secret", "uuid", true, "", " XPanel ")
	metadata, err := auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata["node-role"] != "xpanel" || metadata["node_role"] != "xpanel" {
		t.Fatalf("node role metadata = %#v", metadata)
	}
}

func TestAuthHandlerOmitsEmptyAndInvalidNodeRole(t *testing.T) {
	for _, role := range []string{"", "  ", "generic", "router", "XPANEL-BOX"} {
		auth := NewAuthHandler("secret", "uuid", true, "", role)
		metadata, err := auth.GetRequestMetadata(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := metadata["node-role"]; ok {
			t.Fatalf("role %q must not emit node-role: %#v", role, metadata)
		}
		if _, ok := metadata["node_role"]; ok {
			t.Fatalf("role %q must not emit node_role: %#v", role, metadata)
		}
	}
}

func TestAuthHandlerEmitsOpenWrtRole(t *testing.T) {
	auth := NewAuthHandler("secret", "uuid", true, "", "openwrt")
	metadata, err := auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata["node-role"] != "openwrt" || metadata["node_role"] != "openwrt" {
		t.Fatalf("openwrt metadata = %#v", metadata)
	}
}
