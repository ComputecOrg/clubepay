"use client";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-6xl font-bold text-gray-300 mb-4">Ops!</h1>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">
          Algo deu errado
        </h2>
        <p className="text-gray-500 mb-6">
          Ocorreu um erro inesperado. Tente novamente.
        </p>
        <button
          onClick={reset}
          className="px-6 py-3 rounded-lg text-white font-semibold"
          style={{ backgroundColor: "#2a7d6e" }}
        >
          Tentar novamente
        </button>
      </div>
    </div>
  );
}
