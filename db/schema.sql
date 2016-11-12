CREATE TABLE repositories (
  id SERIAL primary key,
  full_name VARCHAR,
            CONSTRAINT proper_full_name
            CHECK (full_name ~* '^([a-z0-9\-_]+)/([a-z0-9\-_]+)$'),
  token     VARCHAR NOT NULL,
  private   BOOLEAN NOT NULL DEFAULT FALSE
)
