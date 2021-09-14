## Worker pool for expression rest api

A console utility that polls the calculator REST service with several goroutines and save results in postgres.

## Installation:
first need to install: https://github.com/borichevskiy/restcalc
and start service.
Then install worker pool:
```go get github.com/borichevskiy/worker-pool```
## Usage:

```go
docker-compose up --build
```
## Contributing:

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

