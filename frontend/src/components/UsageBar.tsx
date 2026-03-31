interface UsageBarProps {
  used: number;
  limit: number;
  period: "daily" | "monthly";
}

export default function UsageBar({ used, limit, period }: UsageBarProps) {
  const percentage = limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
  const periodLabel = period === "daily" ? "hoje" : "este mês";

  return (
    <div className="flex flex-col gap-2">
      <div className="flex justify-between text-sm text-gray-600">
        <span>
          {used}/{limit} {periodLabel}
        </span>
        <span>{Math.round(percentage)}%</span>
      </div>
      <div className="h-3 w-full rounded-full bg-gray-200 overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-300"
          style={{
            width: `${percentage}%`,
            backgroundColor: "var(--color-primary)",
          }}
        />
      </div>
    </div>
  );
}
