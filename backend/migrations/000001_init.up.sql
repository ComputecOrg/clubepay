CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'owner',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE businesses (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT UNIQUE NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    segment VARCHAR(100) NOT NULL DEFAULT 'cafeteria',
    address TEXT,
    logo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE plans (
    id BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_cents BIGINT NOT NULL,
    limit_type VARCHAR(10) NOT NULL CHECK (limit_type IN ('daily', 'monthly')),
    limit_count INT NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    plan_id BIGINT NOT NULL REFERENCES plans(id),
    subscriber_id BIGINT NOT NULL REFERENCES users(id),
    business_id BIGINT NOT NULL REFERENCES businesses(id),
    psp_subscription_id VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    period_end TIMESTAMPTZ,
    grace_deadline TIMESTAMPTZ,
    referred_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE usages (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id),
    validated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id),
    referred_id BIGINT NOT NULL REFERENCES users(id),
    business_id BIGINT NOT NULL REFERENCES businesses(id),
    code VARCHAR(20) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_subscriptions_business_id ON subscriptions(business_id);
CREATE INDEX idx_subscriptions_subscriber_id ON subscriptions(subscriber_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_psp_id ON subscriptions(psp_subscription_id);
CREATE INDEX idx_usages_subscription_id ON usages(subscription_id);
CREATE INDEX idx_usages_validated_at ON usages(validated_at);
CREATE INDEX idx_plans_business_id ON plans(business_id);
CREATE INDEX idx_referrals_code ON referrals(code);
CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
