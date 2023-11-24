# local-clos
CLOS network simulation on a single linux machine.

It is intended just to study the behavior of CLOS networks in a controlled environment.

## Requirements
* Go
* sudo (for netns)

## Usage
### common part
```
sudo ./make_netns.sh
go build 
```

### terminal(for active)
```
source tmp
s1 ./local-clos active
```

### terminal(for passive)
```
source tmp
l1 ./local-clos passive
```

## for the debug purpose
### tcpdump 
```
tcpdump -v -K 'tcp port 179 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0) '
```
