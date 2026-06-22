CREATE TABLE IF NOT EXISTS links (
    code         TEXT        NOT NULL,
    original_url TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT links_pkey PRIMARY KEY (code),
    CONSTRAINT links_original_url_key UNIQUE (original_url)
);
