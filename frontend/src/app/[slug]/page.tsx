import PlanCard from "@/components/PlanCard";

const API_URL = process.env.API_URL || "http://localhost:8080";

interface Business {
  id: string;
  name: string;
  segment: string;
  address: string;
  slug: string;
  subscriber_count: number;
}

interface Plan {
  id: string;
  name: string;
  description: string;
  price_cents: number;
  limit_type: "daily" | "monthly";
  limit_value: number;
}

async function getBusiness(slug: string): Promise<Business | null> {
  try {
    const res = await fetch(`${API_URL}/api/public/business/${slug}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

async function getPlans(slug: string): Promise<Plan[]> {
  try {
    const res = await fetch(`${API_URL}/api/public/plans/${slug}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return [];
    const data = await res.json();
    return data.plans || data;
  } catch {
    return [];
  }
}

interface PageProps {
  params: Promise<{ slug: string }>;
}

export default async function BusinessLandingPage({ params }: PageProps) {
  const { slug } = await params;
  const [business, plans] = await Promise.all([
    getBusiness(slug),
    getPlans(slug),
  ]);

  if (!business) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Clube não encontrado</h1>
          <p className="mt-2 text-gray-500">
            Este link não existe ou o clube foi desativado.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-2xl px-4 py-16">
        <div className="text-center mb-10">
          <div
            className="inline-block rounded-full px-4 py-1 text-sm font-medium text-white mb-4 bg-primary"
          >
            {business.segment}
          </div>
          <h1 className="text-4xl font-bold text-gray-900">{business.name}</h1>
          {business.address && (
            <p className="mt-2 text-gray-500">{business.address}</p>
          )}
          {business.subscriber_count > 0 && (
            <p className="mt-4 text-gray-600">
              <span className="font-semibold text-gray-900">
                {business.subscriber_count}
              </span>{" "}
              pessoas já assinam
            </p>
          )}
        </div>

        {plans.length === 0 ? (
          <p className="text-center text-gray-500">
            Nenhum plano disponível no momento.
          </p>
        ) : (
          <div className="flex flex-col gap-4">
            <h2 className="text-xl font-semibold text-gray-700">
              Escolha seu plano
            </h2>
            {plans.map((plan) => (
              <PlanCard key={plan.id} plan={plan} slug={slug} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
