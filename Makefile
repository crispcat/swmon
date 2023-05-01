clean:
	./make/make_clean.sh

install:
	./make/make_install.sh

uninstall:
	./make/make_uninstall.sh

clean-install: uninstall clean install

release:
	./make/make_release.sh