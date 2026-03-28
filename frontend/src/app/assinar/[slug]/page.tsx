"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useParams, useSearchParams } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken, setToken } from "@/lib/auth";

interface Plan {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  limit_type: "daily" | "monthly";
  limit_value: number;
}

interface PlansResponse {
  plans: Plan[];
}

interface SubscribeResponse {
  subscription_id: string;
  pix_qr_code?: string;
}

interface AuthResponse {
  token: string;
}

function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

export default function AssinarPage() {
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const slug = params.slug as string;
  const planIdFromQuery = searchParams.get("plan") ?? "";

  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [selectedPlanId, setSelectedPlanId] = useState(planIdFromQuery);
  const [cpf, setCpf] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  // Register-subscriber form fields
  const [subEmail, setSubEmail] = useState("");
  const [subPassword, setSubPassword] = useState("");
  const [subName, setSubName] = useState("");
  const [subPhone, setSubPhone] = useState("");

  const fetchPlans = useCallback(async () => {
    try {
      const data = await api.get<Plan[] | PlansResponse>(
        `/api/public/plans/${slug}`
      );
      const planList = Array.isArray(data)
        ? data
        : (data as PlansResponse).plans ?? [];
      setPlans(planList);
      if (!selectedPlanId && planList.length === 1) {
        setSelectedPlanId(planList[0].id);
      }
    } catch {
      setError("Erro ao carregar planos.");
    } finally {
      setLoading(false);
    }
  }, [slug, selectedPlanId]);

  useEffect(() => {
    const token = getToken();
    setIsLoggedIn(!!token);
    fetchPlans();
  }, [fetchPlans]);

  async function handleSubscribeLoggedIn(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const token = getToken();
    if (!token) return;

    try {
      await api.post<SubscribeResponse>(
        "/api/subscribe",
        { plan_id: selectedPlanId, cpf, business_slug: slug },
        token
      );
      setSuccess(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Erro ao assinar. Tente novamente.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRegisterAndSubscribe(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    try {
      const authData = await api.post<AuthResponse>(
        "/api/auth/register-subscriber",
        {
          email: subEmail,
          password: subPassword,
          name: subName,
          phone: subPhone,
        }
      );
      setToken(authData.token);

      await api.post<SubscribeResponse>(
        "/api/subscribe",
        { plan_id: selectedPlanId, business_slug: slug },
        authData.token
      );
      setSuccess(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Erro ao criar conta e assinar. Tente novamente.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-green-50 px-4">
        <div className="text-center">
          <div className="text-6xl mb-4">🎉</div>
          <h1 className="text-2xl font-bold text-gray-900">
            Assinatura realizada!
          </h1>
          <p className="mt-2 text-gray-500">
            Você já pode começar a usar seu clube.
          </p>
          <button
            onClick={() => router.push("/meu-plano")}
            className="mt-6 rounded-xl px-6 py-3 font-semibold text-white transition-opacity hover:opacity-90"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            Ver meu plano
          </button>
        </div>
      </div>
    );
  }

  const selectedPlan = plans.find((p) => p.id === selectedPlanId);

  return (
    <div className="min-h-screen bg-gray-50 px-4 py-8">
      <div className="mx-auto max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">
            Assinar clube
          </h1>
          {selectedPlan && (
            <p className="mt-1 text-gray-500">
              {selectedPlan.name} —{" "}
              {formatPrice(selectedPlan.price_cents)}/mês
            </p>
          )}
        </div>

        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700 mb-4">
            {error}
          </div>
        )}

        {plans.length > 1 && (
          <div className="mb-6 flex flex-col gap-2">
            <label className="text-sm font-medium text-gray-700">
              Escolha o plano
            </label>
            <select
              value={selectedPlanId}
              onChange={(e) => setSelectedPlanId(e.target.value)}
              className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
            >
              <option value="">Selecione...</option>
              {plans.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} — {formatPrice(p.price_cents)}/mês
                </option>
              ))}
            </select>
          </div>
        )}

        {isLoggedIn ? (
          <form
            onSubmit={handleSubscribeLoggedIn}
            className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-4"
          >
            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700" htmlFor="cpf">
                CPF
              </label>
              <input
                id="cpf"
                type="text"
                required
                value={cpf}
                onChange={(e) => setCpf(e.target.value)}
                className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="000.000.000-00"
                maxLength={14}
              />
            </div>

            <button
              type="submit"
              disabled={submitting || !selectedPlanId}
              className="w-full rounded-xl py-3 font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60"
              style={{ backgroundColor: "#2a7d6e" }}
            >
              {submitting ? "Assinando..." : "Assinar"}
            </button>
          </form>
        ) : (
          <form
            onSubmit={handleRegisterAndSubscribe}
            className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-4"
          >
            <p className="text-sm text-gray-500">
              Crie sua conta para assinar o clube
            </p>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700" htmlFor="subName">
                Nome completo
              </label>
              <input
                id="subName"
                type="text"
                required
                value={subName}
                onChange={(e) => setSubName(e.target.value)}
                className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="Maria Silva"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700" htmlFor="subEmail">
                E-mail
              </label>
              <input
                id="subEmail"
                type="email"
                required
                value={subEmail}
                onChange={(e) => setSubEmail(e.target.value)}
                className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="seu@email.com"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label
                className="text-sm font-medium text-gray-700"
                htmlFor="subPhone"
              >
                Telefone
              </label>
              <input
                id="subPhone"
                type="tel"
                required
                value={subPhone}
                onChange={(e) => setSubPhone(e.target.value)}
                className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="(11) 99999-9999"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label
                className="text-sm font-medium text-gray-700"
                htmlFor="subPassword"
              >
                Senha
              </label>
              <input
                id="subPassword"
                type="password"
                required
                minLength={8}
                value={subPassword}
                onChange={(e) => setSubPassword(e.target.value)}
                className="rounded-lg border border-gray-300 px-3 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-[#2a7d6e]"
                placeholder="Mínimo 8 caracteres"
              />
            </div>

            <button
              type="submit"
              disabled={submitting || !selectedPlanId}
              className="w-full rounded-xl py-3 font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60"
              style={{ backgroundColor: "#2a7d6e" }}
            >
              {submitting ? "Criando conta..." : "Criar conta e assinar"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
