# Docker Registry Cleaner
ローカルRegistryのイメージの削除が面倒だから作ってみた

## Setup
```sh
go get -u github.com/lightstaff/drcleaner
```

## Usage
- tags
```sh
drcleaner tags <target image name>
```
- rm
```sh
drcleaner rm <target image name> <optional: target tag>
```
