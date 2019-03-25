SUCCESS = file_mock_success.csv

################################
########## CSV READER ##########
################################

test-csv_reader:
	-@docker-compose down
	-@docker-compose up -d db_test
	@sleep 5
	-@$ (cd ./cmd/csvreader/test && RUNMODE=TEST go test -count 1)
	-@docker-compose down

run-csv_reader:
	-@docker-compose up -d db
	@sleep 5
	@$ (cd ./cmd/csvreader && go build)
	-@./cmd/csvreader/csvreader $(file)



############################################
############## CRM INTEGRATOR ##############
############################################
test-crm_integrator:
	-@docker-compose down
	-@docker-compose up -d db_test
	@sleep 5
	-@$ (cd ./cmd/crmintegrator/test && RUNMODE=TEST go test -count 1)
	-@docker-compose down

run-crm_integrator:
	-@docker-compose up -d db
	@sleep 5
	@$ (cd ./cmd/crmintegrator && go build)
	-@./cmd/crmintegrator/crmintegrator $(table)