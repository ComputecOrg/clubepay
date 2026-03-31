# AssinaPix — Brand Guideline

> Versão 1.0 | Março 2026

---

## 1. Identidade

**AssinaPix** é a plataforma que permite qualquer negócio físico criar um clube de assinatura e cobrar via Pix recorrente — em 5 minutos.

**Conceito do nome:**
- **Assina** → assinatura, recorrência, o cliente "assina" o clube
- **Pix** → pagamento instantâneo, identidade brasileira, modernidade

**Tom da marca:** Quente, acessível, direto ao ponto. Fala com o dono de padaria ou cafeteria que quer tecnologia sem complicação.

---

## 2. Logo

### Arquivos disponíveis

| Arquivo | Uso |
|---|---|
| `public/brand/logo.svg` | Logo horizontal — uso principal |
| `public/brand/logo-dark.svg` | Logo horizontal — fundo escuro |
| `public/brand/logo-icon.svg` | Ícone quadrado — favicon, avatar, app icon |

### Símbolo (ícone)

Um **diamante** em coral-laranja com um **checkmark** branco interno.

- **Diamante** → referência ao símbolo do Pix (Banco Central do Brasil)
- **Checkmark** → "assinar", confirmar, validar o uso

### Wordmark

- `Assina` em preto (#1A1A18), peso 700
- `Pix` em coral (#E8622A), peso 700
- Ponto separador em teal (#16A394) entre as duas palavras
- Fonte: **Satoshi** (headline) com fallback para system-ui

### Zona de proteção

Mínimo de **8px** de espaço ao redor do logo em qualquer direção.

### O que NÃO fazer

- ❌ Não alterar as cores do logo
- ❌ Não distorcer proporções
- ❌ Não usar o logo sobre fundos com baixo contraste
- ❌ Não recriar o logo em outras fontes

---

## 3. Paleta de Cores

### Primária — Coral Orange

| Token CSS | Hex | Uso |
|---|---|---|
| `--color-primary` | `#E8622A` | Botões CTA, links ativos, destaques principais |
| `--color-primary-dark` | `#BF4A1A` | Hover/active de botões primários |
| `--color-primary-light` | `#FDE8DE` | Fundos de badge, alertas informativos leves |

### Accent — Pix Teal

| Token CSS | Hex | Uso |
|---|---|---|
| `--color-accent` | `#16A394` | Badges "ativo", confirmações, ícones de pagamento |
| `--color-accent-dark` | `#0D7A6E` | Hover de elementos accent |
| `--color-accent-light` | `#D4F5F0` | Fundos suaves para status de sucesso |

### Neutros

| Token CSS | Hex | Uso |
|---|---|---|
| `--color-background` | `#FAFAF8` | Background principal da app |
| `--color-surface` | `#FFFFFF` | Cards, modais, inputs |
| `--color-border` | `#E8E8E5` | Bordas de cards e inputs |
| `--color-muted` | `#F4F4F2` | Fundo de seções secundárias |
| `--color-muted-fg` | `#737370` | Textos secundários, placeholders |
| `--color-foreground` | `#1A1A18` | Texto principal |

### Semânticos

| Token CSS | Hex | Uso |
|---|---|---|
| `--color-success` | `#22C55E` | Confirmação de pagamento, assinatura ativa |
| `--color-warning` | `#F59E0B` | Alertas de dunning, pagamento pendente |
| `--color-danger` | `#EF4444` | Erros, cancelamentos, bloqueios |

---

## 4. Tipografia

### Headline — Satoshi

Fonte geométrica moderna, excelente para títulos e CTAs.

**Instalação no Next.js:**
```bash
# Satoshi não está no Google Fonts — hospedar via fontsource:
npm install @fontsource/satoshi
```

```tsx
// app/layout.tsx
import '@fontsource/satoshi/400.css';
import '@fontsource/satoshi/500.css';
import '@fontsource/satoshi/700.css';
import '@fontsource/satoshi/900.css';
```

**Hierarquia:**
| Elemento | Tamanho | Peso | Font |
|---|---|---|---|
| H1 (hero) | 36–48px | 900 | Satoshi |
| H2 (seção) | 28–32px | 700 | Satoshi |
| H3 (card title) | 20–24px | 700 | Satoshi |
| Body | 16px | 400 | system-ui |
| Small / label | 14px | 500 | system-ui |
| Micro / caption | 12px | 400 | system-ui |

### Body — system-ui

Usa a fonte nativa do sistema operacional. Rápido, legível, nativo.

---

## 5. Componentes — Guia Rápido

### Botão Primário (CTA)
```tsx
<button className="bg-primary hover:bg-primary-dark text-white font-bold px-6 py-3 rounded-xl">
  Assinar agora
</button>
```

### Botão Accent
```tsx
<button className="bg-accent hover:bg-accent-dark text-white font-bold px-6 py-3 rounded-xl">
  Ver planos
</button>
```

### Badge de status ativo
```tsx
<span className="bg-accent-light text-accent-dark text-sm font-medium px-3 py-1 rounded-full">
  Ativo
</span>
```

### Badge de aviso
```tsx
<span className="bg-primary-light text-primary-dark text-sm font-medium px-3 py-1 rounded-full">
  Pagamento pendente
</span>
```

---

## 6. Tom de Voz

- **Direto:** "Crie seu clube em 5 minutos."
- **Caloroso:** Fala de igual para igual com o dono do negócio.
- **Brasileiro:** Pix, clube, assinatura — termos que o público conhece.
- **Sem jargão técnico:** Nunca "API", "webhook", "SaaS" no front-end.

**Exemplos:**
- ✅ "Seus clientes pagam toda mês no Pix, automático."
- ✅ "Crie planos, defina o limite de uso e pronto."
- ❌ "Configure sua subscription via integração PSP."

---

## 7. Referências dos Arquivos

```
clubepay/
├── brand/
│   └── BRAND_GUIDELINE.md       ← este arquivo
└── frontend/
    ├── public/brand/
    │   ├── logo.svg              ← logo horizontal (light)
    │   ├── logo-dark.svg         ← logo horizontal (dark bg)
    │   └── logo-icon.svg         ← ícone quadrado
    ├── src/
    │   ├── app/globals.css       ← tokens CSS (@theme)
    │   └── lib/brand.ts          ← tokens TypeScript
```

---

*AssinaPix Brand v1.0 — Gerado por Pix (Design Agent)*
