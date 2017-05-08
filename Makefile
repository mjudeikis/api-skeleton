build: build-binary ocp-build

build-binary:
	cd 6/ && CGO_ENABLED=0 GOOS=linux go build -o api .

ocp-build:
	oc start-build go-api --from-file=6/ --follow

ocp-set:
	oc project mj ; \
	oc new-build --binary --name go-api ; \
	oc new-build --binary --name java-api ;\
	oc new-build --binary --name python-api

ocp-deploy:
	oc new-app go-api ; \
	oc expose service go-api ; \

maven-build:
	rm -rf java/clean/* ; \
	cd java/gs-spring-boot/initial && mvn clean install

java-build: maven-build	
	cp java/gs-spring-boot/initial/target/gs-spring-boot-0.1.0.jar java/clean/ ;\
	cp java/Dockerfile java/clean ;\
	oc start-build java-api --from-file=java/clean --follow ;\
	oc new-app java-api ; \
	oc expose service java-api 
	
python-build:
	oc start-build python-api --from-file=python/ --follow ;\
	oc new-app python-api  ; \
	oc expose service python-api 

hit-hard-java:
	ab -c 50 -n 1000 http://java-api-mj.apps.192.168.20.187.xip.io/

hit-hard-img:
	 ab -c 50 -n 1000 http://go-api-mj.apps.192.168.20.187.xip.io/api/v1/
hit-hard-health:
	ab -c 50 -n 1000 http://go-api-mj.apps.192.168.20.187.xip.io/health
	