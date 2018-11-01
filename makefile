all: copy

copy:
	rm -rf fwatch/vendor/github.com/seankhliao/igtools/goinsta
	mkdir -p fwatch/vendor/github.com/seankhliao/igtools
	cp -r goinsta fwatch/vendor/github.com/seankhliao/igtools

