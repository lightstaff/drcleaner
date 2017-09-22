# Docker Registry Cleaner
ローカルRegistryのイメージの削除が面倒だから作ってみた

## Setup
```sh
go get -u github.com/lightstaff/drcleaner
```

## Usage

### Command Example
```sh
drcleaner -U=<user regitry url> -T={<target tag1>,<target tag2>} <target image>
```

### Options
- *-url* or *-U* is Tareget URL (Default: localhost:5000)
- *-tags* or *-T* is Target tags (If not set delete all)
- *-verbose* or *-V* is show varbose
