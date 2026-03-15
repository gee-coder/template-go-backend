package service

import (
	"context"
	"testing"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
)

func TestSMSVerificationServiceSendAndVerify(t *testing.T) {
	cache := &fakeCacheStore{}
	svc, err := NewSMSVerificationService(config.SMSConfig{
		Provider: "mock",
		CodeTTL:  5 * time.Minute,
		Cooldown: 30 * time.Second,
		Mock: config.MockSMSConfig{
			RevealCode: true,
			FixedCode:  "123456",
		},
	}, cache, nil)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	payload, err := svc.SendCode(context.Background(), SendSMSCodeInput{
		Phone:   "18800001111",
		Purpose: "login",
	})
	if err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
	if payload.DebugCode != "123456" {
		t.Fatalf("expected mock code to be exposed, got %q", payload.DebugCode)
	}

	if err := svc.VerifyCode(context.Background(), VerifySMSCodeInput{
		Phone:   "18800001111",
		Purpose: "login",
		Code:    "123456",
	}); err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

func TestSMSVerificationServiceCooldown(t *testing.T) {
	cache := &fakeCacheStore{}
	svc, err := NewSMSVerificationService(config.SMSConfig{
		Provider: "mock",
		CodeTTL:  5 * time.Minute,
		Cooldown: time.Minute,
		Mock: config.MockSMSConfig{
			FixedCode: "654321",
		},
	}, cache, nil)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if _, err := svc.SendCode(context.Background(), SendSMSCodeInput{
		Phone:   "18800002222",
		Purpose: "register",
	}); err != nil {
		t.Fatalf("unexpected first send error: %v", err)
	}

	if _, err := svc.SendCode(context.Background(), SendSMSCodeInput{
		Phone:   "18800002222",
		Purpose: "register",
	}); err == nil {
		t.Fatalf("expected cooldown error on repeated send")
	}
}
