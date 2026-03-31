"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken } from "@/lib/auth";

type ValidateState = "idle" | "loading" | "success" | "limit" | "error";

interface ValidateResponse {
  plan_name: string;
  used: number;
  limit: number;
  period: string;
}

export default function ValidarPage() {
  const router = useRouter();
  const params = useParams();
  const slug = params.slug as string;

  const [state, setState] = useState<ValidateState>("idle");
  const [result, setResult] = useState<ValidateResponse | null>(null);
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.push(`/login`);
    }
  }, [router]);

  async function handleValidate() {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    setState("loading");
    setErrorMsg("");

    try {
      const data = await api.post<ValidateResponse>(
        "/api/validate-usage",
        { business_slug: slug },
        token
      );
      setResult(data);
      setState("success");
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 429 || err.message.toLowerCase().includes("limite")) {
          setState("limit");
        } else if (err.status === 404 || err.message.toLowerCase().includes("assinatura")) {
          setErrorMsg("Você não possui assinatura ativa neste clube.");
          setState("error");
        } else {
          setErrorMsg(err.message);
          setState("error");
        }
      } else {
        setErrorMsg("Erro ao validar uso. Tente novamente.");
        setState("error");
      }
    }
  }

  if (state === "success" && result) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-green-500 text-white px-4">
        <div className="text-center">
          <div className="text-7xl mb-4">✅</div>
          <h1 className="text-3xl font-bold">Uso validado!</h1>
          <p className="mt-2 text-xl opacity-90">{result.plan_name}</p>
          <p className="mt-4 text-lg opacity-80">
            {result.used}/{result.limit}{" "}
            {result.period === "daily" ? "hoje" : "este mês"}
          </p>
          <button
            onClick={() => {
              setState("idle");
              setResult(null);
            }}
            className="mt-8 rounded-xl border-2 border-white px-6 py-3 font-semibold transition-opacity hover:opacity-80"
          >
            Voltar
          </button>
        </div>
      </div>
    );
  }

  if (state === "limit") {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-orange-400 text-white px-4">
        <div className="text-center">
          <div className="text-7xl mb-4">⏳</div>
          <h1 className="text-3xl font-bold">Limite atingido</h1>
          <p className="mt-4 text-xl opacity-90">Volte amanhã!</p>
          <button
            onClick={() => setState("idle")}
            className="mt-8 rounded-xl border-2 border-white px-6 py-3 font-semibold transition-opacity hover:opacity-80"
          >
            Voltar
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm text-center flex flex-col gap-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Validar uso</h1>
          <p className="mt-1 text-gray-500 text-sm">
            Confirme seu uso no clube
          </p>
        </div>

        {state === "error" && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
            {errorMsg}
          </div>
        )}

        <button
          onClick={handleValidate}
          disabled={state === "loading"}
          className="w-full rounded-2xl py-5 text-xl font-bold text-white transition-opacity hover:opacity-90 disabled:opacity-60 shadow-md bg-primary"
        >
          {state === "loading" ? "Validando..." : "Confirmar uso"}
        </button>
      </div>
    </div>
  );
}
