package email

import "fmt"

const htmlWrapper = `<html><body style="font-family:system-ui,sans-serif;color:#333;max-width:600px;margin:0 auto;padding:20px">
<div style="background:#2a7d6e;color:white;padding:20px;text-align:center;border-radius:8px 8px 0 0">
<h1 style="margin:0;font-size:24px">ClubePay</h1>
</div>
<div style="padding:20px;border:1px solid #e5e7eb;border-top:none;border-radius:0 0 8px 8px">
%s
</div>
<p style="color:#9ca3af;font-size:12px;text-align:center;margin-top:16px">ClubePay — Clube de assinatura para seu negocio</p>
</body></html>`

func WelcomeEmail(name, planName, businessName string) (string, string) {
	subject := "Bem-vindo ao ClubePay!"
	content := fmt.Sprintf(`<h2>Ola %s!</h2>
<p>Sua assinatura do plano <strong>%s</strong> no <strong>%s</strong> foi criada com sucesso.</p>
<p>Agora voce pode aproveitar todos os beneficios do seu plano.</p>
<p>Obrigado por assinar!</p>`, name, planName, businessName)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func PaymentConfirmedEmail(name, planName, amount string) (string, string) {
	subject := "Pagamento confirmado - ClubePay"
	content := fmt.Sprintf(`<h2>Pagamento confirmado!</h2>
<p>Ola %s, seu pagamento de <strong>%s</strong> para o plano <strong>%s</strong> foi confirmado.</p>
<p>Sua assinatura continua ativa. Aproveite!</p>`, name, amount, planName)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func SubscriptionCancelledEmail(name, planName, validUntil string) (string, string) {
	subject := "Assinatura cancelada - ClubePay"
	content := fmt.Sprintf(`<h2>Assinatura cancelada</h2>
<p>Ola %s, sua assinatura do plano <strong>%s</strong> foi cancelada.</p>
<p>Voce ainda pode usar o servico ate <strong>%s</strong>.</p>
<p>Sentiremos sua falta!</p>`, name, planName, validUntil)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func PasswordResetEmail(name, resetURL string) (string, string) {
	subject := "Redefinir senha - ClubePay"
	content := fmt.Sprintf(`<h2>Redefinir senha</h2>
<p>Ola %s, recebemos um pedido para redefinir sua senha.</p>
<p><a href="%s" style="display:inline-block;background:#2a7d6e;color:white;padding:12px 24px;text-decoration:none;border-radius:6px">Redefinir minha senha</a></p>
<p style="color:#6b7280;font-size:14px">Este link expira em 1 hora. Se voce nao solicitou, ignore este email.</p>`, name, resetURL)
	return subject, fmt.Sprintf(htmlWrapper, content)
}

func GraceBlockedEmail(name string) (string, string) {
	subject := "ClubePay - Assinatura bloqueada"
	content := fmt.Sprintf(`<h2>Assinatura bloqueada</h2>
<p>Ola %s, sua assinatura foi bloqueada por falta de pagamento.</p>
<p>Por favor, regularize seu pagamento para continuar usando o servico.</p>
<p>Equipe ClubePay</p>`, name)
	return subject, fmt.Sprintf(htmlWrapper, content)
}
