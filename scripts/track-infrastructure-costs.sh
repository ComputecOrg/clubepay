#!/bin/bash
# ClubePay Infrastructure Cost Tracking Script
# Usage: ./track-infrastructure-costs.sh [--month YYYY-MM-DD]
#
# This script manually updates monthly_costs for infrastructure expenses.
# Run this daily/weekly to keep spending alerts accurate.

set -e

# Configuration
MONTH="${1:-$(date +%Y-%m-01)}"
DB_USER="${DB_USER:-clubepay}"
DB_PASS="${DB_PASS:-clubepay}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-clubepay_dev}"

# Cost in cents (multiply by 100)
# Claude Max 20x Plan: $200/month
CLAUDE_API_COST=$((200 * 100))  # $200 = 20000 cents

# Coolify VPS (Hostinger KVM1): R$12,99/month
# Using 1 USD = 5 BRL conversion
VPS_COST=$((1299 * 5))  # R$12,99 = 6495 cents

# Domain (.com.br): ~R$3/month
DOMAIN_COST=$((3 * 100))  # R$3 = 300 cents

# Total Monthly Infrastructure Cost
TOTAL_INFRASTRUCTURE=$((CLAUDE_API_COST + VPS_COST + DOMAIN_COST))

echo "========================================"
echo "ClubePay Infrastructure Cost Tracking"
echo "========================================"
echo "Month: $MONTH"
echo ""
echo "Costs:"
echo "  Claude API:    \$200    (20000 cents)"
echo "  VPS (KVM1):    R\$12,99 (6495 cents @ 1 USD = 5 BRL)"
echo "  Domain:        R\$3     (300 cents)"
echo "  ---"
echo "  Total:         ~R\$1.0k (${TOTAL_INFRASTRUCTURE} cents)"
echo ""

# Verify environment
if [ -z "$DATABASE_URL" ] && [ -z "$DB_PASS" ]; then
    echo "ERROR: DATABASE_URL not set and DB_PASS not provided"
    echo "Usage:"
    echo "  export DATABASE_URL='postgres://user:pass@localhost:5432/clubepay_dev'"
    echo "  ./track-infrastructure-costs.sh"
    exit 1
fi

# Build connection string
if [ -n "$DATABASE_URL" ]; then
    PSQL_CONN="$DATABASE_URL"
else
    PSQL_CONN="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
fi

echo "Connecting to database..."
echo ""

# Get first business (for testing/MVP)
BUSINESS_ID=$(psql "$PSQL_CONN" -t -c "SELECT id FROM businesses LIMIT 1;" 2>/dev/null)

if [ -z "$BUSINESS_ID" ]; then
    echo "ERROR: No businesses found in database"
    echo "Please create a business first or check database connection"
    exit 1
fi

echo "Using Business ID: $BUSINESS_ID"
echo ""

# Update monthly_costs for this business and month
echo "Updating monthly_costs..."
psql "$PSQL_CONN" <<EOF
-- Ensure monthly_costs record exists
INSERT INTO monthly_costs (
    business_id,
    month,
    infrastructure_cost_cents,
    total_cost_cents,
    monthly_budget_cents
)
VALUES (
    $BUSINESS_ID,
    '$MONTH'::date,
    $TOTAL_INFRASTRUCTURE,
    $TOTAL_INFRASTRUCTURE,
    500000  -- Default $5000 budget
)
ON CONFLICT (business_id, month)
DO UPDATE SET
    infrastructure_cost_cents = $TOTAL_INFRASTRUCTURE,
    total_cost_cents = $TOTAL_INFRASTRUCTURE,
    updated_at = NOW();

-- Show current spending status
SELECT
    business_id,
    month,
    infrastructure_cost_cents,
    total_cost_cents,
    monthly_budget_cents,
    ROUND(100.0 * total_cost_cents / monthly_budget_cents, 1) as spending_percent,
    CASE
        WHEN 100.0 * total_cost_cents / monthly_budget_cents >= 95 THEN '🔴 CRÍTICO (95%+)'
        WHEN 100.0 * total_cost_cents / monthly_budget_cents >= 80 THEN '🟡 AVISO (80%+)'
        ELSE '🟢 Normal'
    END as status
FROM monthly_costs
WHERE business_id = $BUSINESS_ID AND month = '$MONTH'::date;
EOF

echo ""
echo "✅ Cost tracking updated successfully"
echo ""
echo "Next steps:"
echo "  1. Check dashboard: GET /api/owner/spending/status"
echo "  2. Monitor alerts: Check /api/owner/spending/alerts"
echo "  3. Update monthly: Run this script again for next month"
echo ""
