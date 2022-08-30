all:
	go build .
install:
	mkdir -p $(DESTDIR)/usr/bin
	cp switchman $(DESTDIR)/usr/bin
	chmod 755 $(DESTDIR)/usr/bin/switchman
clean:  
	rm -f switchman
