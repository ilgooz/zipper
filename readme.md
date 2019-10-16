# zipper
Media downloader and zipper with REST API interface.

Zip stream starts immediately, media downloaded one by one and encoded in zip as zip stream is consumed.

# Usage
## Running the Service
* Install Docker and execute start script (`./start.sh`).
* Service will start listening at `:9191` on your host machine.

## Using the Service
There are two options to use the service:

### Use with GET /zip
* This method is easier to test this service with browsers.
* Simply paste following to browser's address bar:
```
http://localhost:9191/zip?files=[{"url":"https://media.giphy.com/media/3oz8xD0xvAJ5FCk7Di/giphy.gif", "filename":"pic001.gif"}]
```
* Customize JSON data under `files` query string as you like.

### Use with POST /zip
* Make a `POST` request to `http://localhost:9191/zip` with JSON payload.
* Example with curl:
```
curl -X POST http://localhost:9191/zip  -H "Content-Type: application/json" -o files.zip -d '[{"url":"https://media.giphy.com/media/3oz8xD0xvAJ5FCk7Di/giphy.gif", "filename":"pic001.gif"}]'
```
* Customize JSON data on the `-d` flag in the example as you like.

### Sample Input
```
[
  {
    url: string;
    filename: string;
  }
]
```

### Sample Output
```
~.zip stream~
```

# Running the Tests
* In order to run tests you need to first install Go and then run `go test ./...` at the root dir.
* Start script (`./start.sh`) already runs these tests inside the Docker container before running the service.