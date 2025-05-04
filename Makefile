dev-run:
	python3 ${GOOGLE_CLOUD_SDK}/bin/dev_appserver.py app.yaml

deploy:
	gcloud app deploy --project=${ESTAMITECH_PROJECT_ID}