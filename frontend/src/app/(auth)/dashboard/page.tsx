"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";
import StatsCard from "@/components/StatsCard";
import BusinessQRCode from "@/components/QRCode";

interface Business {
  id: string;
  name: string;
  slug: string;
  segment: string;
  subscriber_count: number;
}

interface Subscription {
  id: string;
  subscriber_name: string;
  subscriber_email: string;
  plan_name: string;
  status: string;
}

interface SubscriptionsResponse {
  subscriptions: Subscription[];
}

interface BusinessResponse {
  business: Business;
}

function formatMRR(subscriptions: Subscription[]): string {
  // MRR is derived from subscription count — simplified display
  const active = subscriptions.filter((s) => s.status === "active").length;
  return `${active} ativo${active !== 1 ? "s" : ""}`;
}

function statusLabel(status: string): { label: string; color: string } {
  switch (status) {
    case "active":
      return { label: "Ativo", color: "bg-green-100 text-green-700" };
    case "overdue":
      return { label: "Inadimplente", color: "bg-red-100 text-red-700" };
    case "cancelled":
      return { label: "Cancelado", color: "bg-gray-100 text-gray-600" };
    default:
      return { label: status, color: "bg-gray-100 text-gray-600" };
  }
}

export default function DashboardPage() {
  const router = useRouter();
  const [business, setBusiness] = useState<Business | null>(null);
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    try {
      const [bizData, subData] = await Promise.all([
        api.get<BusinessResponse>("/api/business", token),
        api.get<SubscriptionsResponse>("/api/subscriptions", token),
      ]);
      setBusiness(bizData.business);
      setSubscriptions(subData.subscriptions ?? []);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.push("/login");
      } else {
        setError("Erro ao carregar dados.");
      }
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  function handleLogout() {
    clearToken();
    router.push("/login");
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-red-500">{error}</p>
      </div>
    );
  }

  const activeCount = subscriptions.filter((s) => s.status === "active").length;
  const overdueCount = subscriptions.filter((s) => s.status === "overdue").length;
  const adimplencia =
    activeCount + overdueCount > 0
      ? Math.round((activeCount / (activeCount + overdueCount)) * 100)
      : 100;

  const showUpgradeBanner = activeCount >= 12;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-4">
        <div className="mx-auto max-w-4xl flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-gray-900">
              {business?.name ?? "Dashboard"}
            </h1>
            <p className="text-sm text-gray-500">
              clubepay.com/{business?.slug}
            </p>
          </div>
          <button
            onClick={handleLogout}
            className="text-sm text-gray-500 hover:text-gray-800 transition-colors"
          >
            Sair
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-4xl px-4 py-8 flex flex-col gap-6">
        {showUpgradeBanner && (
          <div className="rounded-xl bg-yellow-50 border border-yellow-200 px-4 py-3 flex items-center justify-between">
            <p className="text-sm text-yellow-800 font-medium">
              Você tem {activeCount} assinantes. O plano free suporta até 15.{" "}
              <span className="font-bold">Faça upgrade para continuar crescendo!</span>
            </p>
          </div>
        )}

        {business && (
          <Link
            href={`/validar/${business.slug}`}
            className="block w-full rounded-2xl py-5 text-center text-xl font-bold text-white transition-opacity hover:opacity-90 shadow-md"
            style={{ backgroundColor: "#2a7d6e" }}
          >
            Validar uso
          </Link>
        )}

        <Link
          href="/planos"
          className="block w-full rounded-2xl py-4 text-center text-lg font-bold border-2 transition-colors hover:bg-gray-50"
          style={{ borderColor: "#2a7d6e", color: "#2a7d6e" }}
        >
          Gerenciar planos
        </Link>

        <div className="grid grid-cols-3 gap-4">
          <StatsCard label="Assinantes ativos" value={activeCount} />
          <StatsCard label="MRR" value={formatMRR(subscriptions)} />
          <StatsCard label="Adimplência" value={`${adimplencia}%`} />
        </div>

        {business && <BusinessQRCode slug={business.slug} />}

        <div className="bg-white rounded-2xl border border-gray-200 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="font-semibold text-gray-800">Assinantes</h2>
          </div>

          {subscriptions.length === 0 ? (
            <div className="px-6 py-8 text-center text-gray-500">
              Nenhum assinante ainda.
            </div>
          ) : (
            <ul className="divide-y divide-gray-100">
              {subscriptions.map((sub) => {
                const badge = statusLabel(sub.status);
                return (
                  <li
                    key={sub.id}
                    className="px-6 py-4 flex items-center justify-between"
                  >
                    <div>
                      <p className="font-medium text-gray-900">
                        {sub.subscriber_name}
                      </p>
                      <p className="text-sm text-gray-500">{sub.plan_name}</p>
                    </div>
                    <span
                      className={`rounded-full px-3 py-1 text-xs font-medium ${badge.color}`}
                    >
                      {badge.label}
                    </span>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </main>
    </div>
  );
}
