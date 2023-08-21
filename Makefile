build:
	mkdir build
	go build
	mv ./grave ./build/grave

usrinstall: build
	install -Dt ~/.local/bin -m 755 ./build/grave

install: build
	install -Dt /usr/local/bin -m 755 ./build/grave

clean:
	rm -rf ./build/
