all: clean build run


.SILENT: build clean run

build:
	cd server; \
	go build -o ../bin/genredetector; \
	cd .. 

clean:
	rm -f ./bin/*; \
	
run:
	./bin/genredetector