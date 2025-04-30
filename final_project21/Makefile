execute: deploy
	python3 execute.py

deploy: build
	python3 deploy.py

build:
	chmod +x build.sh
	./build.sh

start_client:
	/build/client_five

report:
	pdflatex --output-directory=../ ../report.tex