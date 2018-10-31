all: copy

copy:
	rm -rf fwatch/vendor/goinsta
	cp -r goinsta fwatch/vendor

