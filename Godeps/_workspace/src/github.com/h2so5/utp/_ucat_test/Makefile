test: ucat gexpect
	go test -timeout 15s -v

ucat: libutp/ucat-static

gexpect:
	go get github.com/ThomasRooney/gexpect

libutp/ucat-static: libutp/Makefile
	cd libutp; make

libutp/Makefile:
	git clone https://github.com/bittorrent/libutp.git

clean:
	rm -rf libutp

.PHONY: clean all
