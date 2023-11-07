#!/bin/sh

# Remove the network namespace if it already exists
ip netns del ns1
ip netns del ns2
# Create a network namespace
ip netns add ns1
ip netns add ns2
# Create a veth pair
ip link add veth1 type veth peer name veth2
# Add the peer interfaces to the namespaces
ip link set veth1 netns ns1
ip link set veth2 netns ns2
# Assign IP addresses
ip netns exec ns1 ip addr add 192.168.0.1/24 dev veth1
ip netns exec ns2 ip addr add 192.168.0.2/24 dev veth2
# Bring up the interfaces
ip netns exec ns1 ip link set dev lo up
ip netns exec ns1 ip link set dev veth1 up
ip netns exec ns2 ip link set dev lo up
ip netns exec ns2 ip link set dev veth2 up

# Alias for command in each namespace
alias ns1="sudo ip netns exec ns1"
alias ns2="sudo ip netns exec ns2"
