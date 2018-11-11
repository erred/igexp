build:
	rm -rf vendor/github.com/seankhliao/igtools/goinsta
	mkdir -p vendor/github.com/seankhliao/igtools
	cp -r ../goinsta vendor/github.com/seankhliao/igtools

deploy:
	gcloud alpha functions deploy mkeep \
		--entry-point Mkeep \
		--env-vars-file env.yaml \
		--retry \
		--runtime go111 \
		--source https://source.developers.google.com/projects/com-seankhliao/repos/igtools/revisions/master/paths/mkeep \
		--timeout=540s \
		--trigger-topic igtools-mkeep

external-trigger:
	gcloud alpha functions call mkeep --data '{"data":"eyJtb2RlIjowfQo="}'

reset:
	gcloud alpha functions delete mkeep
	gcloud pubsub topics delete igtools-mkeep
	gcloud pubsub topics create igtools-mkeep
