-- +goose Up
-- +goose StatementBegin

-- Insert test users (using realistic Telegram user IDs)
INSERT INTO users (id, username, first_name, last_name) VALUES
                                                            (123456789, 'anna_trader', 'Anna', 'Tselikova'),
                                                            (987654321, 'crypto_bob', 'Bob', 'Johnson'),
                                                            (555666777, 'maria_investor', 'Maria', 'Garcia'),
                                                            (111222333, 'test_user', 'Test', 'User'),
                                                            (444555666, 'inactive_user', 'Inactive', 'Account');

-- Insert currency rates (current approximate rates in minor units)
-- Base currency is RUB, so rates show how many kopecks = 1 unit of foreign currency
INSERT INTO currency_rates (currency, rate_minor_units, base_currency, source) VALUES
                                                                                   ('USD', 9250, 'RUB', 'CBR'),  -- 1 USD = 92.50 RUB = 9250 kopecks
                                                                                   ('EUR', 10100, 'RUB', 'CBR'), -- 1 EUR = 101.00 RUB = 10100 kopecks
                                                                                   ('GBP', 11750, 'RUB', 'CBR'), -- 1 GBP = 117.50 RUB = 11750 kopecks
                                                                                   ('CNY', 1280, 'RUB', 'CBR'),  -- 1 CNY = 12.80 RUB = 1280 kopecks
                                                                                   ('KZT', 19, 'RUB', 'CBR'),    -- 1 KZT = 0.19 RUB = 19 kopecks
                                                                                   ('JPY', 62, 'RUB', 'CBR'),    -- 1 JPY = 0.62 RUB = 62 kopecks
                                                                                   ('RSD', 86, 'RUB', 'CBR'),    -- 1 RSD = 0.86 RUB = 86 kopecks
                                                                                   ('XBT', 380000000, 'RUB', 'Binance'); -- 1 BTC = 3,800,000 RUB = 380,000,000 kopecks

-- Insert saving accounts
INSERT INTO saving_accounts (user_id, name, amount_minor_units, interest_rate_basis_points, expiration_date, currency) VALUES
-- Anna's accounts
(123456789, 'Emergency Fund', 50000000, 450, '2025-12-31', 'RUB'),  -- 500,000 RUB at 4.5%
(123456789, 'USD Savings', 500000, 275, '2026-06-30', 'USD'),       -- 5,000 USD at 2.75%
(123456789, 'Euro Vacation Fund', 250000, 180, NULL, 'EUR'),        -- 2,500 EUR at 1.8%

-- Bob's accounts
(987654321, 'High Yield Savings', 1500000, 525, '2025-12-15', 'USD'), -- 15,000 USD at 5.25%
(987654321, 'RUB Account', 100000000, 800, '2026-03-01', 'RUB'),       -- 1,000,000 RUB at 8%

-- Maria's accounts
(555666777, 'Ahorro Principal', 300000000, 650, '2025-09-30', 'RUB'),  -- 3,000,000 RUB at 6.5%
(555666777, 'Euro Savings', 100000, 200, NULL, 'EUR');                 -- 1,000 EUR at 2%

-- Insert deposits (fixed-term investments)
INSERT INTO deposits (user_id, name, amount_minor_units, interest_rate_basis_points, expiration_date, currency) VALUES
-- Anna's deposits
(123456789, 'Sber 12-month Deposit', 200000000, 750, '2025-12-12', 'RUB'), -- 2,000,000 RUB at 7.5%
(123456789, 'USD Term Deposit', 1000000, 425, '2025-11-15', 'USD'),        -- 10,000 USD at 4.25%

-- Bob's deposits
(987654321, 'Tinkoff Deposit', 500000000, 825, '2025-08-20', 'RUB'),       -- 5,000,000 RUB at 8.25%
(987654321, 'Corporate Bond Fund', 250000, 380, '2026-01-10', 'USD'),      -- 2,500 USD at 3.8%

-- Maria's deposits
(555666777, 'Fixed Deposit 6M', 150000000, 700, '2025-09-12', 'RUB'),      -- 1,500,000 RUB at 7%
(555666777, 'Euro Bond', 500000, 250, '2026-02-28', 'EUR');                -- 5,000 EUR at 2.5%

-- Insert brokerage accounts
INSERT INTO brokerage_accounts (user_id, name, amount_minor_units, currency, broker_name, account_type) VALUES
-- Anna's brokerage
(123456789, 'Tinkoff Investments', 75000000, 'RUB', 'Tinkoff', 'regular'),     -- 750,000 RUB
(123456789, 'IIS Account', 40000000, 'RUB', 'Sber', 'iis'),                   -- 400,000 RUB
(123456789, 'Interactive Brokers', 1200000, 'USD', 'IBKR', 'regular'),        -- 12,000 USD

-- Bob's brokerage
(987654321, 'Crypto Portfolio', 5000000, 'XBT', 'Binance', 'regular'),        -- 0.05 BTC (5,000,000 satoshi)
(987654321, 'Stock Portfolio', 2500000, 'USD', 'Fidelity', 'ira'),            -- 25,000 USD
(987654321, 'Russian Stocks', 300000000, 'RUB', 'Tinkoff', 'regular'),        -- 3,000,000 RUB

-- Maria's brokerage
(555666777, 'Investment Account', 180000000, 'RUB', 'VTB', 'regular'),        -- 1,800,000 RUB
(555666777, 'Pension Fund', 80000000, 'RUB', 'Sber', 'iis'),                  -- 800,000 RUB

-- Test user brokerage
(111222333, 'Test Brokerage', 100000000, 'RUB', 'Test Broker', 'regular');    -- 1,000,000 RUB

-- Insert cash holdings
INSERT INTO cash_holdings (user_id, name, amount_minor_units, currency) VALUES
-- Anna's cash
(123456789, 'Wallet RUB', 5000000, 'RUB'),     -- 50,000 RUB
(123456789, 'Travel Cash USD', 50000, 'USD'),   -- 500 USD
(123456789, 'Emergency EUR', 20000, 'EUR'),     -- 200 EUR
(123456789, 'Petty Cash KZT', 5000000, 'KZT'),  -- 50,000 KZT

-- Bob's cash
(987654321, 'Wallet', 3000000, 'RUB'),          -- 30,000 RUB
(987654321, 'USD Cash', 100000, 'USD'),         -- 1,000 USD
(987654321, 'Bitcoin Wallet', 1000000, 'XBT'),  -- 0.01 BTC (1,000,000 satoshi)

-- Maria's cash
(555666777, 'Efectivo RUB', 8000000, 'RUB'),    -- 80,000 RUB
(555666777, 'Euros en Casa', 50000, 'EUR'),     -- 500 EUR
(555666777, 'Dolares', 25000, 'USD'),           -- 250 USD

-- Test user cash
(111222333, 'Test Cash RUB', 10000000, 'RUB'),  -- 100,000 RUB
(111222333, 'Test Cash USD', 100000, 'USD'),    -- 1,000 USD
(111222333, 'Test Cash JPY', 50000, 'JPY'),     -- 50,000 JPY (no minor unit)

-- Inactive user (minimal data)
(444555666, 'Old Wallet', 1000000, 'RUB');      -- 10,000 RUB

-- Add some variety with different currencies for testing
INSERT INTO cash_holdings (user_id, name, amount_minor_units, currency) VALUES
                                                                            (123456789, 'Serbian Dinars', 1000000, 'RSD'),  -- 10,000 RSD
                                                                            (987654321, 'Chinese Yuan', 500000, 'CNY'),     -- 5,000 CNY
                                                                            (555666777, 'British Pounds', 100000, 'GBP');   -- 1,000 GBP

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Delete test data in reverse order (respecting foreign key constraints)
DELETE FROM cash_holdings WHERE user_id IN (123456789, 987654321, 555666777, 111222333, 444555666);
DELETE FROM brokerage_accounts WHERE user_id IN (123456789, 987654321, 555666777, 111222333, 444555666);
DELETE FROM deposits WHERE user_id IN (123456789, 987654321, 555666777, 111222333, 444555666);
DELETE FROM saving_accounts WHERE user_id IN (123456789, 987654321, 555666777, 111222333, 444555666);
DELETE FROM currency_rates WHERE currency IN ('USD', 'EUR', 'GBP', 'CNY', 'KZT', 'JPY', 'RSD', 'XBT');
DELETE FROM users WHERE id IN (123456789, 987654321, 555666777, 111222333, 444555666);

-- +goose StatementEnd