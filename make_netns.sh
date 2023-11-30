#!/bin/sh

# Remove the network namespace if it already exists
ip netns del spine-1
ip netns del spine-2
ip netns del leaf-1
ip netns del leaf-2
# Create a network namespace
ip netns add spine-1
ip netns add spine-2
ip netns add leaf-1
ip netns add leaf-2
# Create a veth pair
ip link add veth-s1-l1-s type veth peer name veth-s1-l1-l
ip link add veth-s1-l2-s type veth peer name veth-s1-l2-l
ip link add veth-s2-l1-s type veth peer name veth-s2-l1-l
ip link add veth-s2-l2-s type veth peer name veth-s2-l2-l
# Add the peer interfaces to the namespaces
# いい感じにこの後の設定書いて
ip link set veth-s1-l1-s netns spine-1
ip link set veth-s1-l1-l netns leaf-1
ip link set veth-s1-l2-s netns spine-1
ip link set veth-s1-l2-l netns leaf-2
ip link set veth-s2-l1-s netns spine-2
ip link set veth-s2-l1-l netns leaf-1
ip link set veth-s2-l2-s netns spine-2
ip link set veth-s2-l2-l netns leaf-2
# Assign IP addresses
# いい感じにこの後の設定書いて
ip netns exec spine-1 ip addr add 10.1.1.1/24 dev veth-s1-l1-s
ip netns exec leaf-1 ip addr add 10.1.1.2/24 dev veth-s1-l1-l
ip netns exec spine-1 ip addr add 10.1.2.1/24 dev veth-s1-l2-s
ip netns exec leaf-2 ip addr add 10.1.2.2/24 dev veth-s1-l2-l
ip netns exec spine-2 ip addr add 10.2.1.1/24 dev veth-s2-l1-s
ip netns exec leaf-1 ip addr add 10.2.1.2/24 dev veth-s2-l1-l
ip netns exec spine-2 ip addr add 10.2.2.1/24 dev veth-s2-l2-s
ip netns exec leaf-2 ip addr add 10.2.2.2/24 dev veth-s2-l2-l
# Bring up the interfaces
ip netns exec spine-1 ip link set dev veth-s1-l1-s up
ip netns exec leaf-1 ip link set dev veth-s1-l1-l up
ip netns exec spine-1 ip link set dev veth-s1-l2-s up
ip netns exec leaf-2 ip link set dev veth-s1-l2-l up
ip netns exec spine-2 ip link set dev veth-s2-l1-s up
ip netns exec leaf-1 ip link set dev veth-s2-l1-l up
ip netns exec spine-2 ip link set dev veth-s2-l2-s up
ip netns exec leaf-2 ip link set dev veth-s2-l2-l up

ip netns exec spine-1 ip link set dev lo up
ip netns exec spine-2 ip link set dev lo up
ip netns exec leaf-1 ip link set dev lo up
ip netns exec leaf-2 ip link set dev lo up


ip netns exec spine-1 sysctl -w net.ipv4.conf.veth-s1-l1-s.arp_accept=1
ip netns exec spine-1 sysctl -w net.ipv4.conf.veth-s1-l2-s.arp_accept=1
ip netns exec spine-2 sysctl -w net.ipv4.conf.veth-s2-l1-s.arp_accept=1
ip netns exec spine-2 sysctl -w net.ipv4.conf.veth-s2-l2-s.arp_accept=1
ip netns exec leaf-1 sysctl -w net.ipv4.conf.veth-s1-l1-l.arp_accept=1  
ip netns exec leaf-1 sysctl -w net.ipv4.conf.veth-s2-l1-l.arp_accept=1
ip netns exec leaf-2 sysctl -w net.ipv4.conf.veth-s1-l2-l.arp_accept=1
ip netns exec leaf-2 sysctl -w net.ipv4.conf.veth-s2-l2-l.arp_accept=1

echo 'alias s1="sudo ip netns exec spine-1"' > shell_alias
echo 'alias s2="sudo ip netns exec spine-2"' >> shell_alias
echo 'alias l1="sudo ip netns exec leaf-1"' >> shell_alias
echo 'alias l2="sudo ip netns exec leaf-2"' >> shell_alias