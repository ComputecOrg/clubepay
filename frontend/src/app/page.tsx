import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
      <div className="text-center max-w-lg flex flex-col gap-6">
        <div>
          <h1 className="text-5xl font-bold tracking-tight text-primary">
            ClubePay
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
  );
}
