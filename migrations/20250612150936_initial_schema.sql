-- +goose Up
-- +goose StatementBegin
CREATE TYPE account_type_enum AS ENUM (
    'saving',
    'deposit',
    'brokerage',
    'cash'
);

CREATE TYPE currency_enum AS ENUM (
    'RUB',  -- Russian Ruble (kopecks)
    'USD',  -- US Dollar (cents)
    'EUR',  -- Euro (cents)
    'GBP',  -- British Pound (pence)
    'JPY',  -- Japanese Yen (already minor unit)
    'CNY',  -- Chinese Yuan (fen)
    'RSD',  -- Serbian Dinar (para)
    'XBT',  -- Bitcoin (satoshi)
    'KZT'   -- Kazakhstani Tenge (tiyn)
);

CREATE TYPE brokerage_type_enum AS ENUM (
    'regular',
    'iis',      -- Individual Investment Account (Russia)
    'iis3',     -- Individual Investment Account Type 3 (Russia)
    'ira',      -- Individual Retirement Account (US)
    'margin'
);

CREATE TABLE currency_minor_units (
                                      currency currency_enum PRIMARY KEY,
                                      minor_unit_name VARCHAR(50) NOT NULL,
                                      units_per_major INTEGER NOT NULL,
                                      symbol VARCHAR(10) NOT NULL,
                                      description TEXT
);

INSERT INTO currency_minor_units VALUES
                                     ('RUB', 'kopeck', 100, '₽', 'Russian Ruble'),
                                     ('USD', 'cent', 100, '$', 'US Dollar'),
                                     ('EUR', 'cent', 100, '€', 'Euro'),
                                     ('GBP', 'penny', 100, '£', 'British Pound'),
                                     ('JPY', 'yen', 1, '¥', 'Japanese Yen (no minor unit)'),
                                     ('CNY', 'fen', 100, '¥', 'Chinese Yuan'),
                                     ('RSD', 'para', 100, 'RSD', 'Serbian Dinar'),
                                     ('XBT', 'satoshi', 100000000, '₿', 'Bitcoin'),
                                     ('KZT', 'tiyn', 100, '₸', 'Kazakhstani Tenge');

CREATE TABLE users (
                       id BIGINT PRIMARY KEY,
                       username VARCHAR(255),
                       first_name VARCHAR(255),
                       last_name VARCHAR(255),
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE saving_accounts (
                                 id BIGSERIAL PRIMARY KEY,
                                 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                 name VARCHAR(255) NOT NULL,
                                 amount_minor_units BIGINT NOT NULL DEFAULT 0, -- Amount in minor units (kopecks, cents, etc.)
                                 interest_rate_basis_points INTEGER, -- e.g., 525 for 5.25% (1 basis point = 0.01%)
                                 expiration_date DATE,
                                 currency currency_enum NOT NULL DEFAULT 'RUB',
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                 updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE deposits (
                          id BIGSERIAL PRIMARY KEY,
                          user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          name VARCHAR(255) NOT NULL,
                          amount_minor_units BIGINT NOT NULL DEFAULT 0, -- Amount in minor units
                          interest_rate_basis_points INTEGER, -- e.g., 525 for 5.25% (1 basis point = 0.01%)
                          expiration_date DATE,
                          currency currency_enum NOT NULL DEFAULT 'RUB',
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE brokerage_accounts (
                                    id BIGSERIAL PRIMARY KEY,
                                    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                    name VARCHAR(255) NOT NULL,
                                    amount_minor_units BIGINT NOT NULL DEFAULT 0, -- Amount in minor units
                                    currency currency_enum NOT NULL DEFAULT 'RUB',
                                    broker_name VARCHAR(255), -- e.g., "Tinkoff", "Sber"
                                    account_type brokerage_type_enum DEFAULT 'regular',
                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cash_holdings (
                               id BIGSERIAL PRIMARY KEY,
                               user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               name VARCHAR(255) NOT NULL, -- e.g., "Wallet USD", "Safe RUB"
                               amount_minor_units BIGINT NOT NULL DEFAULT 0, -- Amount in minor units
                               currency currency_enum NOT NULL,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE currency_rates (
                                id BIGSERIAL PRIMARY KEY,
                                currency currency_enum NOT NULL,
                                rate_minor_units BIGINT NOT NULL, -- Exchange rate in minor units of base currency
                                base_currency currency_enum NOT NULL DEFAULT 'RUB',
                                source VARCHAR(50) DEFAULT 'CBR', -- Central Bank of Russia
                                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                UNIQUE(currency, base_currency)
);

-- Helper functions
CREATE OR REPLACE FUNCTION major_to_minor_units(
    amount_major DECIMAL,
    curr currency_enum
) RETURNS BIGINT AS $$
DECLARE
multiplier INTEGER;
BEGIN
SELECT units_per_major INTO multiplier
FROM currency_minor_units
WHERE currency = curr;

RETURN (amount_major * multiplier)::BIGINT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION minor_to_major_units(
    amount_minor BIGINT,
    curr currency_enum
) RETURNS DECIMAL AS $$
DECLARE
divisor INTEGER;
BEGIN
SELECT units_per_major INTO divisor
FROM currency_minor_units
WHERE currency = curr;

RETURN amount_minor::DECIMAL / divisor;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION basis_points_to_percentage(
    basis_points INTEGER
) RETURNS DECIMAL AS $$
BEGIN
RETURN basis_points::DECIMAL / 10000;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION percentage_to_basis_points(
    percentage DECIMAL
) RETURNS INTEGER AS $$
BEGIN
RETURN (percentage * 10000)::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Indexes
CREATE INDEX idx_saving_accounts_user_id ON saving_accounts(user_id);
CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_brokerage_accounts_user_id ON brokerage_accounts(user_id);
CREATE INDEX idx_cash_holdings_user_id ON cash_holdings(user_id);
CREATE INDEX idx_currency_rates_currency ON currency_rates(currency);
CREATE INDEX idx_currency_rates_updated ON currency_rates(updated_at);
CREATE INDEX idx_currency_rates_base ON currency_rates(base_currency);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_currency_rates_base;
DROP INDEX IF EXISTS idx_currency_rates_updated;
DROP INDEX IF EXISTS idx_currency_rates_currency;
DROP INDEX IF EXISTS idx_cash_holdings_user_id;
DROP INDEX IF EXISTS idx_brokerage_accounts_user_id;
DROP INDEX IF EXISTS idx_deposits_user_id;
DROP INDEX IF EXISTS idx_saving_accounts_user_id;

-- Drop functions
DROP FUNCTION IF EXISTS percentage_to_basis_points(DECIMAL);
DROP FUNCTION IF EXISTS basis_points_to_percentage(INTEGER);
DROP FUNCTION IF EXISTS minor_to_major_units(BIGINT, currency_enum);
DROP FUNCTION IF EXISTS major_to_minor_units(DECIMAL, currency_enum);

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS currency_rates;
DROP TABLE IF EXISTS cash_holdings;
DROP TABLE IF EXISTS brokerage_accounts;
DROP TABLE IF EXISTS deposits;
DROP TABLE IF EXISTS saving_accounts;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS currency_minor_units;

-- Drop enums (in reverse order)
DROP TYPE IF EXISTS brokerage_type_enum;
DROP TYPE IF EXISTS currency_enum;
DROP TYPE IF EXISTS account_type_enum;
-- +goose StatementEnd