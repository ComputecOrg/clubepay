"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";


interface MyPlanResponse {
  plan: {
    id: number;
    name: string;
    description: string;
    price_cents: number;
    limit_type: "daily" | "monthly";
    limit_count: number;
  };
  business: {
    id: number;
    name: string;
    slug: string;
  };
  subscription: {
    id: number;
    status: string;
    period_end: string | null;
  };
}

interface ReferralCodeResponse {
  code: string;
}

function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("pt-BR");
}

function statusLabel(status: string): string {
  switch (status) {
    case "active":
      return "Ativo";
    case "overdue":
      return "Inadimplente";
    case "cancelled":
      return "Cancelado";
    default:
      return status;
  }
}

export default function MeuPlanoPage() {
  const router = useRouter();
  const [plan, setPlan] = useState<MyPlanResponse | null>(null);
  const [referralCode, setReferralCode] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [cancelLoading, setCancelLoading] = useState(false);

  const fetchData = useCallback(async () => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    try {
      const [planData, referralData] = await Promise.all([
        api.get<MyPlanResponse>("/api/my-plan"),
        api.get<ReferralCodeResponse>("/api/my-referral-code"),
      ]);
      setPlan(planData);
      setReferralCode(referralData.code);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.push("/login");
      } else {
        setError("Erro ao carregar dados do plano.");
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  async function handleCancel() {
    const confirmed = window.confirm(
      "Tem certeza que deseja cancelar sua assinatura? Você continuará com acesso até o fim do período pago."
    );
    if (!confirmed) return;

    const token = getToken();
    if (!token) return;

    setCancelLoading(true);
    try {
      await api.post("/api/cancel", {});
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Erro ao cancelar assinatura.");
      }
    } finally {
      setCancelLoading(false);
    }
  }

  function handleShareReferral() {
    const text = `Use meu código ${referralCode} no ClubePay e ganhe 10% de desconto!`;
    if (navigator.share) {
      navigator.share({ text });
    } else {
      navigator.clipboard.writeText(text);
      alert("Código copiado para a área de transferência!");
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  if (error && !plan) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-red-500">{error}</p>
      </div>
    );
  }

  if (!plan) return null;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-4">
        <div className="mx-auto max-w-lg">
          <h1 className="text-xl font-bold text-primary">
            ClubePay
          </h1>
        </div>
      </header>

      <main className="mx-auto max-w-lg px-4 py-8 flex flex-col gap-6">
        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-4">
          <div>
            <p className="text-sm text-gray-500">{plan.business.name}</p>
            <h2 className="text-2xl font-bold text-gray-900">{plan.plan.name}</h2>
          </div>

          <div className="flex flex-col gap-2 text-sm text-gray-600">
            <div className="flex justify-between">
              <span>Valor</span>
              <span className="font-medium text-gray-900">
                {formatPrice(plan.plan.price_cents)}/mes
              </span>
            </div>
            <div className="flex justify-between">
              <span>Limite</span>
              <span className="font-medium text-gray-900">
                {plan.plan.limit_count}x por {plan.plan.limit_type === "daily" ? "dia" : "mes"}
              </span>
            </div>
            {plan.subscription.period_end && (
              <div className="flex justify-between">
                <span>Proxima cobranca</span>
                <span className="font-medium text-gray-900">
                  {formatDate(plan.subscription.period_end)}
                </span>
              </div>
            )}
            <div className="flex justify-between">
              <span>Status</span>
              <span
                className={`font-medium ${
                  plan.subscription.status === "active"
                    ? "text-green-600"
                    : plan.subscription.status === "overdue"
                    ? "text-red-600"
                    : "text-gray-600"
                }`}
              >
                {statusLabel(plan.subscription.status)}
              </span>
            </div>
          </div>
        </div>

        {referralCode && (
          <div className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-3">
            <div>
              <h3 className="font-semibold text-gray-900">Indicar amigo</h3>
              <p className="text-sm text-gray-500">
                Você e seu amigo ganham 10% de desconto
              </p>
            </div>
            <div className="flex items-center gap-3">
              <code className="flex-1 rounded-lg bg-gray-100 px-3 py-2 text-sm font-mono text-gray-800">
                {referralCode}
              </code>
              <button
                onClick={handleShareReferral}
                className="rounded-xl px-4 py-2 text-sm font-semibold text-white transition-opacity hover:opacity-90 bg-accent"
              >
                Indicar amigo
              </button>
            </div>
          </div>
        )}

        <div className="text-center">
          <button
            onClick={handleCancel}
            disabled={cancelLoading}
            className="text-sm text-gray-400 hover:text-red-500 transition-colors disabled:opacity-60"
          >
            {cancelLoading ? "Cancelando..." : "Cancelar assinatura"}
          </button>
        </div>
      </main>
    </div>
  );
}
