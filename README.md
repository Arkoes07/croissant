# Croissant

## Getting Started

### Installing

* Clone repository
* Get dependancies
```
go mod vendor
```

### Executing program
* create a secret file `files/secret/secret.json` with JSON structure defines `files/secret/secret-template.json`, and fill it with the correct value
* build
```
go build
```
* run
```
./croissant.exe
```