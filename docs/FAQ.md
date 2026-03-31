# FAQ - ClubePay

## Perguntas Frequentes

### 1. **O que é ClubePay?**
ClubePay é uma ferramenta que permite qualquer negócio físico (cafeterias, padarias, academias, etc.) criar um clube de assinatura e cobrar via Pix recorrente em apenas 5 minutos, usando a integração com Asaas.

### 2. **Quanto custa usar ClubePay?**
ClubePay é **gratuito até 15 assinantes e 1 plano**. Após atingir esses limites, você precisa fazer upgrade para continuar adicionando novos assinantes e planos.

### 3. **Como funciona o pagamento?**
Os clientes pagam via QR Pix mensal (ou no período definido para o plano). O dinheiro entra direto na sua conta Asaas - ClubePay não fica com nada, apenas facilia a cobrança.

### 4. **Quem recebe o dinheiro das assinaturas?**
Você recebe direto no Asaas. ClubePay apenas processa e gerencia os QR códigos de pagamento.

### 5. **Posso ter múltiplos planos?**
Sim! Você pode criar quantos planos quiser. Na versão gratuita, está limitado a 1 plano. Faça upgrade para ter múltiplos planos.

### 6. **Como os clientes se cadastram?**
Os clientes acessam a landing page do seu negócio (incluída no ClubePay) e clicam em "Assinar". Preenchem um formulário rápido e recebem um QR Pix para pagar.

### 7. **O que é "validação de uso"?**
Cada assinante pode validar seu uso (ex: 1 café por dia ou 4 banhos por mês). Ao validar, o app marca que usou parte do seu limite.

### 8. **O que acontece se o pagamento falhar?**
ClubePay oferece 3 dias de carência. Nesse período, o assinante continua usando normalmente. Se não pagar nos 3 dias, o acesso é bloqueado até regularizar.

### 9. **Posso indicar amigos?**
Sim! Todo assinante tem um código de indicação. Quem usa o código recebe 10% de desconto - e você também ganha 10% de desconto. Limite: 3 indicações ativas por assinante.

### 10. **Como funciona o cancelamento?**
O assinante pode cancelar a qualquer momento. O acesso continua até o fim do período já pago (ex: se pagou até dia 30, usa até lá).

### 11. **Quais são as regras de limites de uso?**
Você define por plano:
- **Daily:** Limite por dia (ex: "1 café por dia")
- **Monthly:** Limite por mês (ex: "4 banhos por mês")

### 12. **ClubePay está em segurança?**
Sim! Senhas são criptografadas com bcrypt. Tokens JWT assinam todas as requisições. Webhooks do Asaas são validados via HMAC.

### 13. **Posso usar ClubePay offline?**
Parcialmente. Para validar uso, o assinante precisa estar online (logado). Existe um fallback: se o QR falhar, pode buscar por nome/telefone.

### 14. **Quem gerencia os dados dos meus clientes?**
Você! Os dados ficam no seu banco de dados PostgreSQL. ClubePay apenas facilita a interface e integração com Asaas para pagamentos.

### 15. **Como posso fazer upgrade do plano gratuito?**
Entre em contato pelo formulário "Contato" na landing page. Oferecemos planos custom dependendo do seu volume de assinantes.

---

**Ainda tem dúvidas?** Envie um email para suporte@clubepay.com ou use o chat de suporte na plataforma.
