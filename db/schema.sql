CREATE TABLE repositories (
  id SERIAL primary key,
  owner VARCHAR
  CONSTRAINT proper_owner
            CHECK (owner ~* '^([a-z0-9\-_]+)$')
  NOT NULL,
  name VARCHAR
  CONSTRAINT proper_name
            CHECK (name ~* '^([a-z0-9\-_]+)$')
  NOT NULL,
  private BOOLEAN NOT NULL DEFAULT FALSE
)
