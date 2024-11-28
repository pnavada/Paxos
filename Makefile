# Docker image name
DOCKER_IMAGE=prj4

# Docker targets
docker:
	docker build -t $(DOCKER_IMAGE) .

# Test case targets
run-test1:
	docker-compose -f docker-compose-testcase-1.yml up

run-test2:
	docker-compose -f docker-compose-testcase-2.yml up

stop-test:
	docker-compose -f docker-compose-testcase-1.yml down
	docker-compose -f docker-compose-testcase-2.yml down

docker-clean:
	docker rmi $(DOCKER_IMAGE)
	docker system prune -f