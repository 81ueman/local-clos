# local-clos
CLOS network simulation on a single linux machine.

It is intended just to study the behavior of CLOS networks in a controlled environment.

## Requirements
* Go
* sudo (to use netns, ip route, ip rule)

## Usage
### common part
```
sudo ./make_netns.sh
go build 
```

### terminal(for active)
```
source tmp
s1 ./local-clos -mode=active -as=65000
```

### terminal(for passive)
```
source tmp
l1 ./local-clos -mode=passive -as=65001
```

## for the debug purpose
### tcpdump 
```
tcpdump -v -K 'tcp port 179 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0) '
```

