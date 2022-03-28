all: clean build run


.SILENT: build clean run

build:
	cd server; \
	cp ./config.json ~/.config/genredetector/config.json; \
	go build -o ../bin/genredetector; \
	cd .. 

clean:
	rm -f ./bin/*; \
	
run:
	./bin/genredetector