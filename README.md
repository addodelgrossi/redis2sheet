# Redis 2 Sheet

Redis event to Google Sheet

## Usage

```bash
go run main.go run --spreadsheetID 1b6yWP-uurheUvLgg0PENgYyZmLknqGHkRP9gsuHTBmI --channels=events,copy
``````

## Test

```bash
go test -v -run TestPublishEvent
```

## Download

### Linux

```bash
curl -L https://github.com/addodelgrossi/redis2sheet/releases/download/v0.0.2/redis2sheet_linux-amd64 -o redis2sheet
chmod +x redis2sheet
```
