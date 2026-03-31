-- Table to track monthly costs per business
CREATE TABLE monthly_costs (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id),
    month DATE NOT NULL,
    infrastructure_cost_cents BIGINT NOT NULL DEFAULT 0,
    claude_api_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost_cents BIGINT NOT NULL DEFAULT 0,
    monthly_budget_cents BIGINT NOT NULL DEFAULT 500000, -- $5000 default
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(business_id, month)
);

-- Table to track alerts already sent to avoid duplicates
CREATE TABLE spending_alerts (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id),
    monthly_cost_id BIGINT NOT NULL REFERENCES monthly_costs(id),
    alert_level VARCHAR(10) NOT NULL CHECK (alert_level IN ('warning', 'critical')),
    threshold_percent INT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_monthly_costs_business_month ON monthly_costs(business_id, month);
CREATE INDEX idx_monthly_costs_business_id ON monthly_costs(business_id);
CREATE INDEX idx_spending_alerts_business_id ON spending_alerts(business_id);
CREATE INDEX idx_spending_alerts_monthly_cost_id ON spending_alerts(monthly_cost_id);
