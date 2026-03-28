import Link from "next/link";

interface Plan {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  limit_type: "daily" | "monthly";
  limit_value: number;
}

interface PlanCardProps {
  plan: Plan;
  slug: string;
}

function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  });
}

function formatLimit(limitType: "daily" | "monthly", limitValue: number): string {
  if (limitType === "daily") {
    return `${limitValue}x por dia`;
  }
  return `${limitValue}x por mês`;
}

export default function PlanCard({ plan, slug }: PlanCardProps) {
  return (
    <div className="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h3 className="text-xl font-semibold text-gray-900">{plan.name}</h3>
        {plan.description && (
          <p className="text-gray-500 text-sm">{plan.description}</p>
        )}
      </div>

      <div className="flex flex-col gap-1">
        <span className="text-3xl font-bold text-gray-900">
          {formatPrice(plan.price_cents)}
        </span>
        <span className="text-sm text-gray-500">por mês</span>
      </div>

      <div className="text-sm text-gray-600 bg-gray-50 rounded-lg px-3 py-2 inline-block w-fit">
        {formatLimit(plan.limit_type, plan.limit_value)}
      </div>

      <Link
        href={`/assinar/${slug}?plan=${plan.id}`}
        className="mt-2 block w-full rounded-xl py-3 text-center font-semibold text-white transition-opacity hover:opacity-90"
        style={{ backgroundColor: "#d4a853" }}
      >
        Assinar agora
      </Link>
    </div>
  );
}
