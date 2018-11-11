# stateful example
This example persists tracing information using Consul.

## pre-requisites
### Consul
#### OSX
```bash
brew install consul
consul agent -dev
```
### Honeycomb
The example uses a honeycomb.io account with the write key and datasets passed in via environment variables
```bash
export HONEYCOMB_KEY=REDACTED
export HONEYCOMB_DATASET="my-data-set"
```

## build
```bash
make
```

## run
```bash
./watcher
```