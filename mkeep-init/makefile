deploy:
	gcloud alpha functions deploy test1 --entry-point F --runtime go111 --source https://source.developers.google.com/projects/com-seankhliao/repos/igtools/revisions/master/paths/mkeep-init  --trigger-http

execute:
	gcloud alpha functions call test1
