PACKAGES := $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

test: setupdb
	@echo "Running tests..."
	DB_USER=postgres DB_NAME=bw_hooks_test go test $(PACKAGES)


setupdb:
	@echo "Setting up test database..."
	psql -U postgres -c "DROP DATABASE IF EXISTS bw_hooks_test;"
	psql -U postgres -c "CREATE DATABASE bw_hooks_test;"
	psql -U postgres bw_hooks_test < db/schema.sql
