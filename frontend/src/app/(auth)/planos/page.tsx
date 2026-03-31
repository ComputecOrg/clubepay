"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";

interface Plan {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  limit_type: "daily" | "monthly";
  limit_count: number;
  active: boolean;
}

interface PlansResponse {
  plans: Plan[];
}

function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

export default function PlanosPage() {
  const router = useRouter();
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [showForm, setShowForm] = useState(false);
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    price_cents: "",
    limit_type: "daily" as "daily" | "monthly",
    limit_count: "1",
  });
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState("");

  const fetchPlans = useCallback(async () => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    try {
      const data = await api.get<PlansResponse>("/api/plans");
      setPlans(data.plans ?? []);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.push("/login");
      } else {
        setError("Erro ao carregar planos.");
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    fetchPlans();
  }, [fetchPlans]);

  function resetForm() {
    setFormData({
      name: "",
      description: "",
      price_cents: "",
      limit_type: "daily",
      limit_count: "1",
    });
    setEditingPlan(null);
    setShowForm(false);
    setFormError("");
  }

  function handleEdit(plan: Plan) {
    setEditingPlan(plan);
    setFormData({
      name: plan.name,
      description: plan.description || "",
      price_cents: (plan.price_cents / 100).toString(),
      limit_type: plan.limit_type,
      limit_count: plan.limit_count.toString(),
    });
    setShowForm(true);
    setFormError("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token) return;

    setSubmitting(true);
    setFormError("");

    const priceCents = Math.round(parseFloat(formData.price_cents) * 100);
    if (isNaN(priceCents) || priceCents <= 0) {
      setFormError("Preco invalido.");
      setSubmitting(false);
      return;
    }

    const limitCount = parseInt(formData.limit_count, 10);
    if (isNaN(limitCount) || limitCount <= 0) {
      setFormError("Limite invalido.");
      setSubmitting(false);
      return;
    }

    const payload = {
      name: formData.name,
      description: formData.description,
      price_cents: priceCents,
      limit_type: formData.limit_type,
      limit_count: limitCount,
    };

    try {
      if (editingPlan) {
        await api.put(`/api/plans/${editingPlan.id}`, payload);
      } else {
        await api.post("/api/plans", payload);
      }
      resetForm();
      fetchPlans();
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message);
      } else {
        setFormError("Erro ao salvar plano.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(planId: number) {
    const confirmed = window.confirm("Tem certeza que deseja desativar este plano?");
    if (!confirmed) return;

    const token = getToken();
    if (!token) return;

    try {
      await api.del(`/api/plans/${planId}`);
      fetchPlans();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      }
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-4">
        <div className="mx-auto max-w-2xl flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">Meus Planos</h1>
          <Link
            href="/dashboard"
            className="text-sm font-medium transition-colors hover:opacity-80 text-primary"
          >
            Voltar ao Dashboard
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-2xl px-4 py-8 flex flex-col gap-6">
        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {!showForm && (
          <button
            onClick={() => { resetForm(); setShowForm(true); }}
            className="w-full rounded-2xl py-4 text-lg font-bold text-white transition-opacity hover:opacity-90 shadow-md bg-primary"
          >
            + Criar novo plano
          </button>
        )}

        {showForm && (
          <form onSubmit={handleSubmit} className="bg-white rounded-2xl border border-gray-200 p-6 flex flex-col gap-4">
            <h2 className="font-semibold text-gray-800">
              {editingPlan ? "Editar plano" : "Novo plano"}
            </h2>

            {formError && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
                {formError}
              </div>
            )}

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">Nome</label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                placeholder="Ex: Cafe Diario"
              />
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">Descricao (opcional)</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                placeholder="1 cafe por dia"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1">
                <label className="text-sm font-medium text-gray-700">Preco (R$)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  required
                  value={formData.price_cents}
                  onChange={(e) => setFormData({ ...formData, price_cents: e.target.value })}
                  className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  placeholder="29.90"
                />
              </div>

              <div className="flex flex-col gap-1">
                <label className="text-sm font-medium text-gray-700">Tipo de limite</label>
                <select
                  value={formData.limit_type}
                  onChange={(e) => setFormData({ ...formData, limit_type: e.target.value as "daily" | "monthly" })}
                  className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
                >
                  <option value="daily">Diario</option>
                  <option value="monthly">Mensal</option>
                </select>
              </div>
            </div>

            <div className="flex flex-col gap-1">
              <label className="text-sm font-medium text-gray-700">
                Limite ({formData.limit_type === "daily" ? "usos por dia" : "usos por mes"})
              </label>
              <input
                type="number"
                min="1"
                required
                value={formData.limit_count}
                onChange={(e) => setFormData({ ...formData, limit_count: e.target.value })}
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
              />
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={submitting}
                className="flex-1 rounded-xl py-3 font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60 bg-primary"
              >
                {submitting ? "Salvando..." : editingPlan ? "Atualizar" : "Criar plano"}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="rounded-xl px-6 py-3 font-semibold text-gray-600 bg-gray-100 hover:bg-gray-200 transition-colors"
              >
                Cancelar
              </button>
            </div>
          </form>
        )}

        {plans.length === 0 && !showForm ? (
          <div className="text-center py-12 text-gray-500">
            Nenhum plano criado ainda. Crie seu primeiro plano!
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {plans.map((plan) => (
              <div
                key={plan.id}
                className="bg-white rounded-2xl border border-gray-200 p-6 flex items-center justify-between"
              >
                <div>
                  <h3 className="font-semibold text-gray-900">{plan.name}</h3>
                  {plan.description && (
                    <p className="text-sm text-gray-500">{plan.description}</p>
                  )}
                  <p className="text-sm text-gray-600 mt-1">
                    {formatPrice(plan.price_cents)}/mes — {plan.limit_count}x{" "}
                    {plan.limit_type === "daily" ? "por dia" : "por mes"}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleEdit(plan)}
                    className="rounded-lg px-3 py-1.5 text-sm font-medium border border-gray-300 text-gray-600 hover:bg-gray-50 transition-colors"
                  >
                    Editar
                  </button>
                  <button
                    onClick={() => handleDelete(plan.id)}
                    className="rounded-lg px-3 py-1.5 text-sm font-medium border border-red-200 text-red-600 hover:bg-red-50 transition-colors"
                  >
                    Desativar
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
