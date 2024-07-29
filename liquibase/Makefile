ROOT_PATH = $(shell pwd)

migrate-local:
	docker run --network=host --rm \
		-v $(ROOT_PATH)/migration:/liquibase/changelog liquibase/liquibase \
		--searchPath=/liquibase/changelog/ \
		--changeLogFile=changelog.xml \
		--url="jdbc:postgresql://localhost:5432/liquibase?user=admin&password=admin&currentSchema=public&sslmode=disable" \
		update

migrate-docker:
	docker run --network=postgres_default --rm \
		-v $(ROOT_PATH)/migration:/liquibase/changelog liquibase/liquibase \
		--searchPath=/liquibase/changelog/ \
		--changeLogFile=changelog.xml \
		--url="jdbc:postgresql://pg-projects/liquibase?user=admin&password=admin&currentSchema=public&sslmode=disable" \
		update
