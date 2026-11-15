CREATE TABLE users
(
    id          UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    telegram_id BIGINT       NOT NULL UNIQUE,
    username    VARCHAR(64),
    first_name  VARCHAR(255) NOT NULL,
    last_name   VARCHAR(255),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE books
(
    id          UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    isbn        VARCHAR(13),
    title       VARCHAR(500) NOT NULL,
    author      VARCHAR(255) NOT NULL,
    description TEXT,
    genre       VARCHAR(100),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE book_copies
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    book_id     UUID        NOT NULL REFERENCES books (id),
    owner_id    UUID        NOT NULL REFERENCES users (id),
    condition   VARCHAR(30) NOT NULL,
    daily_price BIGINT      NOT NULL,
    deposit     BIGINT      NOT NULL DEFAULT 0,
    status      VARCHAR(30) NOT NULL DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT book_copies_condition_check CHECK ( condition IN ('new', 'good', 'fair', 'poor')),
    CONSTRAINT book_copies_status_check CHECK ( status IN ('available', 'unavailable', 'archived')),
    CONSTRAINT book_copies_daily_price_check CHECK ( daily_price >= 0 )
);

CREATE INDEX idx_book_copies_book_id ON book_copies (book_id);
CREATE INDEX idx_book_copies_owner_id ON book_copies (owner_id);
CREATE INDEX idx_book_copies_status ON book_copies (status);
