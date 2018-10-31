all: copy

copy:
	rm -rf fwatch/vendor/igtools/goinsta
	mkdir -p fwatch/vendor/igtools
	cp -r goinsta fwatch/vendor/igtools

