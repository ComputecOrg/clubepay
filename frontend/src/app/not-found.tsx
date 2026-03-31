import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-8xl font-bold text-gray-200 mb-4">404</h1>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">
          Pagina nao encontrada
        </h2>
        <p className="text-gray-500 mb-6">
          A pagina que voce procura nao existe ou foi movida.
        </p>
        <Link
          href="/"
          className="inline-block px-6 py-3 rounded-lg text-white font-semibold bg-primary"
        >
          Voltar ao inicio
        </Link>
      </div>
    </div>
  );
}
