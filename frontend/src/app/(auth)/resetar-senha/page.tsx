"use client";

import { Suspense, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";

function ResetarSenhaForm() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  if (!token) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 flex flex-col gap-4 text-center">
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          Link inválido ou expirado.
        </div>
        <p className="text-sm text-gray-500">
          <Link
            href="/esqueci-senha"
            className="font-semibold text-primary"
          >
            Solicitar novo link
          </Link>
        </p>
      </div>
    );
  }

  if (success) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 flex flex-col gap-4 text-center">
        <div className="rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-700">
          Senha atualizada com sucesso!
        </div>
        <p className="text-sm text-gray-500">
          <Link
            href="/login"
            className="font-semibold text-primary"
          >
            Ir para o login
          </Link>
        </p>
      </div>
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (password.length < 8) {
      setError("A senha deve ter no mínimo 8 caracteres.");
      return;
    }

    if (password !== confirmPassword) {
      setError("As senhas não coincidem.");
      return;
    }

    setLoading(true);

    try {
      await api.post("/api/auth/confirm-password-reset", { token, password });
      setSuccess(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Erro ao redefinir senha. Tente novamente.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 flex flex-col gap-4"
    >
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700" htmlFor="password">
          Nova senha
        </label>
        <input
          id="password"
          type="password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
          placeholder="••••••••"
        />
      </div>

      <div className="flex flex-col gap-1">
        <label className="text-sm font-medium text-gray-700" htmlFor="confirmPassword">
          Confirmar senha
        </label>
        <input
          id="confirmPassword"
          type="password"
          required
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
          placeholder="••••••••"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="mt-2 w-full rounded-xl py-3 font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60 bg-primary"
      >
        {loading ? "Salvando..." : "Redefinir senha"}
      </button>

      <p className="text-center text-sm text-gray-500">
        <Link
          href="/login"
          className="font-semibold text-primary"
        >
          Voltar para o login
        </Link>
      </p>
    </form>
  );
}

export default function ResetarSenhaPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-primary">
            ClubePay
          </h1>
          <p className="mt-2 text-gray-500">Redefinir senha</p>
        </div>

        <Suspense
          fallback={
            <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 text-center text-gray-500 text-sm">
              Carregando...
            </div>
          }
        >
          <ResetarSenhaForm />
        </Suspense>
      </div>
    </div>
  );
}
