banner:
	go run ./cmd/banner

dev-run:
	python3 ${GOOGLE_CLOUD_SDK}/bin/dev_appserver.py app.yaml

deploy:
	gcloud config set account ${ESTAMITECH_ACCOUNT_ID}
	gcloud app deploy --project=${ESTAMITECH_PROJECT_ID}