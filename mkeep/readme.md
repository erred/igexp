# mkeep

downloads instagram feed/tags/media

## Development

```
(collection) Users
    |- (doc) User1
    |   - BlacklistFeed: false
    |   - BlacklistStory: false
    |   - BlacklistTag: true
    `- (doc) User2
        - BlacklistFeed: false
        - BlacklistStory: false
        - BlacklistTag: true
(collection) Media
    |- (doc) MediaID01
    |   - User: UserID
    |   - Date: Date
    `- (doc) MediaID02
        - User: UserID
        - Date: Date
```

## Deployment

Cloud Functions Go - Pub/Sub trigger

```
{"Mode":0}
```

### Env Vars

- `BUCKET` - Cloud storage bucket
- `TOPIC` - pubsub topic to use as task queue
- `FIRE` - firestore database

### Bootstrap

- `$BUCKET/mkeep/goinsta.json` : valid [goinsta](https://github.com/ahmdrz/goinsta) export file
