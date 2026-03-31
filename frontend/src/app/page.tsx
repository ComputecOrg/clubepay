import Link from "next/link";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "AssinaPix — Clube de Assinatura via Pix para Negócios Físicos",
  description:
    "Crie seu clube de assinatura Pix em 5 minutos. Cobranças automáticas mensais via Pix recorrente. Ideal para cafeterias, padarias, salões e barbearias.",
  alternates: {
    canonical: "https://assinapix.com.br",
  },
};

const websiteSchema = {
  "@context": "https://schema.org",
  "@type": "WebSite",
  name: "AssinaPix",
  url: "https://assinapix.com.br",
  description:
    "Plataforma de clube de assinatura via Pix recorrente para negócios físicos.",
  potentialAction: {
    "@type": "SearchAction",
    target: {
      "@type": "EntryPoint",
      urlTemplate: "https://assinapix.com.br/blog?q={search_term_string}",
    },
    "query-input": "required name=search_term_string",
  },
};

const softwareSchema = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "AssinaPix",
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web",
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "BRL",
    description: "Gratuito até 15 assinantes",
  },
  description:
    "Crie seu clube de assinatura Pix em 5 minutos. Cobranças automáticas mensais via Pix recorrente para qualquer negócio físico.",
};

export default function Home() {
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(websiteSchema) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareSchema) }}
      />
      <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
        <div className="text-center max-w-lg flex flex-col gap-6">
          <div>
            <h1 className="text-5xl font-bold tracking-tight text-primary">
              AssinaPix
            </h1>
            <p className="mt-3 text-xl text-gray-600">
              Crie seu clube de assinatura em 5 minutos
            </p>
          </div>

          <p className="text-gray-500">
            Cobranças automáticas via Pix. Seus clientes assinam, você foca no
            negócio.
          </p>

          <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
            <Link
              href="/register"
              className="rounded-xl px-8 py-3 text-base font-semibold text-white transition-opacity hover:opacity-90 bg-primary"
            >
              Criar meu clube
            </Link>
            <Link
              href="/login"
              className="rounded-xl border border-gray-300 bg-white px-8 py-3 text-base font-semibold text-gray-700 transition-colors hover:bg-gray-50"
            >
              Já tenho conta
            </Link>
          </div>
        </div>
      </div>
    </>
  );
}
