-- Monthly costs tracking for spending alerts
CREATE TABLE monthly_costs (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    month DATE NOT NULL,
    infrastructure_cost_cents BIGINT NOT NULL DEFAULT 0,
    claude_api_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost_cents BIGINT NOT NULL DEFAULT 0,
    monthly_budget_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(business_id, month)
);

CREATE INDEX idx_monthly_costs_business_id ON monthly_costs(business_id);
CREATE INDEX idx_monthly_costs_month ON monthly_costs(month);

-- Spending alerts for budget threshold notifications
CREATE TABLE spending_alerts (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    monthly_cost_id BIGINT NOT NULL REFERENCES monthly_costs(id) ON DELETE CASCADE,
    alert_level VARCHAR(50) NOT NULL,
    threshold_percent INT NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_spending_alerts_business_id ON spending_alerts(business_id);
CREATE INDEX idx_spending_alerts_monthly_cost_id ON spending_alerts(monthly_cost_id);
CREATE INDEX idx_spending_alerts_sent_at ON spending_alerts(sent_at);
