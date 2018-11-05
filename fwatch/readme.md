# fwatch

keeps tracks of instagram follows

## Development

[goinsta](https://github.com/ahmdrz/goinsta) is forked to fix versioning issues with Go 1.11, vendor using `make` from `..`

## Deployment

Cloud Functions Go - Pub/Sub trigger

### Env Vars

- `BUCKET` - Cloud storage bucket

### Bootstrap

- `$BUCKET/fwatch/goinsta.json` : valid [goinsta](https://github.com/ahmdrz/goinsta) export file
