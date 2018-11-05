# ticker

Uses appengine's cron service to trigger publish a message to pubsub

proxies empty message from /path/to/$pubsub-topic to $pubsub-topic

## Deployment

Remember to deploy both `app.yaml` and `cron.yaml`
