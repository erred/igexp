build:
	rm -rf vendor/github.com/seankhliao/igtools/goinsta
	mkdir -p vendor/github.com/seankhliao/igtools
	cp -r ../goinsta vendor/github.com/seankhliao/igtools

deploy:
	gcloud alpha functions deploy mkeep \
		--entry-point Mkeep \
		--env-vars-file env.yaml \
		--region asia-northeast1 \
		--runtime go111 \
		--source https://source.developers.google.com/projects/com-seankhliao/repos/igtools/revisions/master/paths/mkeep \
		--trigger-topic igtools-mkeep
