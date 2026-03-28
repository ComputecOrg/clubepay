package email_test

import (
	"testing"

	"github.com/clubepay/backend/internal/email"
	"github.com/stretchr/testify/assert"
)

func TestMockSender_RecordsMessages(t *testing.T) {
	mock := &email.MockSender{}

	err := mock.Send("user@test.com", "Assinatura bloqueada", "Sua assinatura foi bloqueada.")
	assert.NoError(t, err)
	assert.Len(t, mock.Sent, 1)
	assert.Equal(t, "user@test.com", mock.Sent[0].To)
	assert.Equal(t, "Assinatura bloqueada", mock.Sent[0].Subject)
}

func TestNewSMTP_ReturnsNilWhenNotConfigured(t *testing.T) {
	sender := email.NewSMTP("", "", "", "")
	assert.Nil(t, sender)
}

func TestNewSMTP_ReturnsSenderWhenConfigured(t *testing.T) {
	sender := email.NewSMTP("smtp.test.com", "587", "user", "pass")
	assert.NotNil(t, sender)
}
