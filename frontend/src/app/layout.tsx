import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://assinapix.com.br"),
  title: {
    default: "AssinaPix — Clube de Assinatura via Pix",
    template: "%s | AssinaPix",
  },
  description:
    "AssinaPix: crie seu clube de assinatura Pix em 5 minutos. Cobranças automáticas via Pix recorrente para cafeterias, padarias, salões e qualquer negócio físico.",
  keywords: [
    "AssinaPix",
    "assina pix",
    "clube de assinatura",
    "Pix recorrente",
    "assinatura mensal",
    "pagamento Pix",
    "clube mensal",
    "cafeteria assinatura",
    "padaria assinatura",
    "negócio físico assinatura",
    "cobrança recorrente Pix",
  ],
  openGraph: {
    type: "website",
    locale: "pt_BR",
    url: "https://assinapix.com.br",
    siteName: "AssinaPix",
    title: "AssinaPix — Clube de Assinatura via Pix",
    description:
      "Crie seu clube de assinatura Pix em 5 minutos. Cobranças automáticas via Pix recorrente para qualquer negócio físico.",
  },
  twitter: {
    card: "summary_large_image",
    title: "AssinaPix — Clube de Assinatura via Pix",
    description:
      "Crie seu clube de assinatura Pix em 5 minutos. Cobranças automáticas via Pix recorrente.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR" className="h-full">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
