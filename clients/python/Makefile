all: gem

generate:
	python -m grpc_tools.protoc -I ../.. -I ../../vendor -I ../../../../.. --python_out=. --grpc_python_out=. ../../pqstream.proto
	git diff --exit-code

deps:
	pip install -r requirements.txt
	pip install -r requirements.dev.txt

lint:
	pylint example.py
