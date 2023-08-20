build:
	mkdir build
	go build
	mv ./grave ./build/grave

usrinstall: build
	cp ./build/grave ~/.bin/

install: build
	sudo cp ./build/grave /bin/

clean:
	rm -rf ./build/
