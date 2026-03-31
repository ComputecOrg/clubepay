package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWelcomeEmail(t *testing.T) {
	subject, body := WelcomeEmail("Joao", "Cafe Premium", "Padaria do Bairro")
	assert.Equal(t, "Bem-vindo ao AssinaPix!", subject)
	assert.Contains(t, body, "Joao")
	assert.Contains(t, body, "Cafe Premium")
	assert.Contains(t, body, "Padaria do Bairro")
	assert.Contains(t, body, "<html")
}

func TestPaymentConfirmedEmail(t *testing.T) {
	subject, body := PaymentConfirmedEmail("Joao", "Cafe Premium", "R$ 29,90")
	assert.Equal(t, "Pagamento confirmado - AssinaPix", subject)
	assert.Contains(t, body, "Joao")
	assert.Contains(t, body, "R$ 29,90")
}

func TestSubscriptionCancelledEmail(t *testing.T) {
	subject, body := SubscriptionCancelledEmail("Joao", "Cafe Premium", "30/04/2026")
	assert.Equal(t, "Assinatura cancelada - AssinaPix", subject)
	assert.Contains(t, body, "Joao")
	assert.Contains(t, body, "30/04/2026")
}

func TestPasswordResetEmail(t *testing.T) {
	subject, body := PasswordResetEmail("Joao", "https://assinapix.com.br/resetar-senha?token=abc123")
	assert.Equal(t, "Redefinir senha - AssinaPix", subject)
	assert.Contains(t, body, "Joao")
	assert.Contains(t, body, "https://assinapix.com.br/resetar-senha?token=abc123")
}

func TestGraceBlockedEmail(t *testing.T) {
	subject, body := GraceBlockedEmail("Joao")
	assert.Equal(t, "AssinaPix - Assinatura bloqueada", subject)
	assert.Contains(t, body, "Joao")
	assert.Contains(t, body, "bloqueada")
}
