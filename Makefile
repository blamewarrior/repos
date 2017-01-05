PACKAGES := $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)

test: setupdb
	@echo "Running tests..."
	DB_USER=postgres DB_NAME=bw_repos_test go test $(PACKAGES) -p 1


setupdb:
	@echo "Setting up test database..."
	psql -U postgres -c "DROP DATABASE IF EXISTS bw_repos_test;"
	psql -U postgres -c "CREATE DATABASE bw_repos_test;"
	psql -U postgres bw_repos_test < db/schema.sql
