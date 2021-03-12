
create-network:
	-docker network create fusionauth_network_stack

build-image:
	docker build -t fa-middleware:latest .

run-fa: create-network
	docker-compose -f docker-compose.fa.yml up -d

run-middleware: create-network
	docker-compose -f docker-compose.yml up -d

run: create-network
	docker-compose -f docker-compose.fa.yml -f docker-compose.yml up -d

down-middleware:
	docker-compose -f docker-compose.yml down

down:
	docker-compose -f docker-compose.fa.yml -f docker-compose.yml down

logs:
	docker-compose -f docker-compose.fa.yml -f docker-compose.yml logs -f
