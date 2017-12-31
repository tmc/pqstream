all: lint

test:
	TEST_PYTHON_INTEGRATION_EXEC="cd clients/python; coverage run example.py" go test -v github.com/tmc/pqstream -run TestServer_Listen/python_integration
	bash -c 'bash <(curl -s https://codecov.io/bash) -F python'

generate:
	python -m grpc_tools.protoc -I ../.. -I ../../vendor -I ../../../../.. --python_out=. --grpc_python_out=. ../../pqstream.proto
	git diff --exit-code

deps:
	pip install -r requirements.txt
	pip install -r requirements.dev.txt

lint:
	pylint example.py
