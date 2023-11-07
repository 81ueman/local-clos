# local-clos
CLOS network simulation on a single linux machine.

It is intended just to study the behavior of CLOS networks in a controlled environment.

## Requirements
* Go
* sudo (for netns)

## Usage
```
sudo ./make_netns.sh
go build 
sudo ./local-clos passive
sudo ./local-clos active
```

