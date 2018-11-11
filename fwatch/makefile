build:
	rm -rf vendor/github.com/seankhliao/igtools/goinsta
	mkdir -p vendor/github.com/seankhliao/igtools
	cp -r ../goinsta vendor/github.com/seankhliao/igtools

deploy:
	gcloud alpha functions deploy fwatch \
		--entry-point Fwatch \
		--env-vars-file env.yaml \
		--retry \
		--runtime go111 \
		--source https://source.developers.google.com/projects/com-seankhliao/repos/igtools/revisions/master/paths/fwatch \
		--trigger-topic igtools-fwatch
