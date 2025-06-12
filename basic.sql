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
    'KZT',  -- Kazakhstani Tenge (tïin)
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
                       id BIGINT PRIMARY KEY,  -- Telegram user ID
                       username VARCHAR(255),  -- Telegram username (optional)
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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



