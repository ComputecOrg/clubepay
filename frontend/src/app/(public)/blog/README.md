# Blog ClubePay

Este é o blog do ClubePay, implementado com Next.js 15+ em modo SSG (Static Site Generation) para máxima performance.

## Estrutura

```
blog/
├── page.tsx           # Listagem de posts (/blog)
├── [slug]/
│   └── page.tsx       # Post individual (/blog/my-post-slug)
├── content/
│   └── *.md          # Posts em Markdown
└── README.md         # Este arquivo
```

## Como Adicionar Posts

### 1. Criar arquivo Markdown

Crie um arquivo `.md` na pasta `content/`:

```bash
touch frontend/src/app/(public)/blog/content/meu-artigo.md
```

### 2. Adicionar frontmatter YAML

Todo post começa com frontmatter YAML:

```yaml
---
title: "Título do Artigo"
date: "2026-04-15"
author: "Nome do Autor"
excerpt: "Resumo breve que aparece na listagem"
tags: ["tag1", "tag2"]
---
```

### 3. Escrever conteúdo em Markdown

```markdown
# Título do Artigo

Conteúdo aqui em Markdown.

## Seção 1

Parágrafo...

### Subseção

Mais conteúdo...
```

### Exemplo Completo

```markdown
---
title: "5 Ideias de Assinatura para sua Cafeteria"
date: "2026-04-01"
author: "Hugo Conteúdo"
excerpt: "Descubra como transformar clientes ocasionais em assinantes recorrentes"
tags: ["assinatura", "cafeteria"]
---

# 5 Ideias de Assinatura para sua Cafeteria

Muitos donos...

## 1. Assinatura de Café Premium

...
```

## Campos de Frontmatter

| Campo | Obrigatório | Descrição |
|-------|-------------|-----------|
| `title` | ✅ | Título do artigo |
| `date` | ✅ | Data de publicação (YYYY-MM-DD) |
| `author` | ⚠️ | Nome do autor (opcional) |
| `excerpt` | ⚠️ | Resumo (aparece na listagem) |
| `tags` | ⚠️ | Array de tags |
| `readingTime` | ⚠️ | Tempo de leitura (calculado automaticamente) |

## URLs

- **Blog Home:** `/blog`
- **Post Individual:** `/blog/{slug}` (ex: `/blog/5-ideias-assinatura-cafeteria`)

### Como o slug é gerado?

O slug é o nome do arquivo **sem** `.md`:

```
Arquivo: 5-ideias-assinatura-cafeteria.md
URL: /blog/5-ideias-assinatura-cafeteria
```

## Build & Deploy

O blog é automaticamente compilado em SSG (Static Site Generation):

```bash
npm run build
# Gera HTML estático pra cada post
```

Posts aparecem automaticamente na listagem ordenados por **data (mais recente primeiro)**.

## SEO

Cada post inclui:
- ✅ Meta tags dinâmicas (title, description)
- ✅ Open Graph (para compartilhamento em redes)
- ✅ Timestamps estruturados
- ✅ Heading hierarchy correta (H1 → H2 → H3)

## Performance

- **SSG:** Posts são compilados em build time (zero latência)
- **Markdown:** Convertido em HTML via `marked` (rápido)
- **Images:** Otimizadas automaticamente pelo Next.js
- **Cache:** Posts cachados indefinidamente (mudam raramente)

## Troubleshooting

### Post não aparece na listagem
- Verifique se o arquivo está em `content/`
- Verifique a sintaxe YAML do frontmatter
- Verifique a data (posts futuros não aparecem)

### Formatação quebrada
- Use Markdown padrão (GitHub Flavored Markdown)
- Verifique indentação (4 espaços ou tabs)
- Teste em https://markdownlivepreview.com/

### Slug incorreto
- Use nomes de arquivo em minúsculas
- Use hífens para espaços (não underscores)
- Exemplo: `meu-artigo-sobre-assinatura.md`

## Integração com Analytics

Ver [CLU-6 Analytics](/CLU/issues/CLU-6) para configurar tracking de blog posts.

## Próximas Features

- [ ] Search de posts
- [ ] Comentários
- [ ] RSS feed
- [ ] Newsletter signup
- [ ] Related posts
