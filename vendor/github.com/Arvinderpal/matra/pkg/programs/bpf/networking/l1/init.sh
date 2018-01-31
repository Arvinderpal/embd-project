#!/bin/bash

# Attach ingress direction:
# ./init.sh /root/go/src/github.com/Arvinderpal/matra/pkg/programs/bpf /var/run/matra/ep1 direct eth0 ingress
# ./init.sh /root/go/src/github.com/Arvinderpal/matra/pkg/programs/bpf /var/run/matra/ep1 direct eth0 ingress,egress

# Remove bpf:
# ./init.sh /root/go/src/github.com/Arvinderpal/matra/pkg/programs/bpf /var/run/matra/ep1 REMOVE

LIB=$1
RUNDIR=$2
# ADDR=$3
# V4ADDR=$4
MODE=$3
# Only set if MODE = "direct" or "lb"
NATIVE_DEV=$4
DIRECTION=$5

DEBUG

MOUNTPOINT="/sys/fs/bpf"

SRC_IN=networking/l1/c/traffic_counter.c
OBJ_OUT=$RUNDIR/traffic_counter.o 
SEC_NAME=from-netdev

set -e
set -x

# Enable JIT
echo 1 > /proc/sys/net/core/bpf_jit_enable

# check to see $MOUNTPOINT is mounted, if not, mount it
if [ $(mount | grep $MOUNTPOINT > /dev/null) ]; then
	mount bpffs $MOUNTPOINT -t bpf
fi

function mac2array()
{
	echo "{0x${1//:/,0x}}"
}
NODE_MAC=$(ip link show $NATIVE_DEV | grep ether | awk '{print $2}')
NODE_MAC="{.addr=$(mac2array $NODE_MAC)}"

################
#  BPF Compile #
################
# This directory was created by the daemon and contains the per container header file
DIR="$PWD/globals" # TODO(awander): not sure we need this
if [ -z "$DEBUG" ]; then
	CLANG_OPTS="-D__NR_CPUS__=$(nproc) -O2 -target bpf -I$RUNDIR -I$DIR -I. -I$LIB -I$LIB/include -DHANDLE_NS -DDEBUG"
else
	CLANG_OPTS="-D__NR_CPUS__=$(nproc) -O2 -target bpf -I$RUNDIR -I$DIR -I. -I$LIB -I$LIB/include -DHANDLE_NS"
fi

# NOTE(awander): note the cleaver way to pass defines to the compiler!
# ID=$(regulus policy get-id $WORLD_ID 2> /dev/null)
# OPTS="-DSECLABEL=${ID} -DPOLICY_MAP=regulus_policy_reserved_${ID}"
clang $CLANG_OPTS $OPTS -DNODE_MAC=${NODE_MAC} -c $LIB/$SRC_IN -o $OBJ_OUT

################
#      tc      #
################
if [ "$MODE" = "direct" ]; then
	if [ -z "$NATIVE_DEV" ]; then
		echo "No device specified for direct mode, ignoring..."
	else
		tc qdisc del dev $NATIVE_DEV clsact 2> /dev/null || true
		tc qdisc add dev $NATIVE_DEV clsact
		if [[ $DIRECTION == "egress" ]]; then
			tc filter add dev $NATIVE_DEV egress bpf da obj $OBJ_OUT sec $SEC_NAME
			sysctl -q -w net.ipv4.conf.$NATIVE_DEV.forwarding=1 #  do we need this?
		elif [[ $DIRECTION == "ingress,egress" ]]; then
			tc filter add dev $NATIVE_DEV ingress bpf da obj $OBJ_OUT sec $SEC_NAME
			tc filter add dev $NATIVE_DEV egress bpf da obj $OBJ_OUT sec $SEC_NAME
			# e.g. tc filter add dev eth1 ingress bpf da obj bpf_g3.o sec from-netdev
			sysctl -q -w net.ipv4.conf.$NATIVE_DEV.forwarding=1 #  do we need this?
		else
			tc filter add dev $NATIVE_DEV ingress bpf da obj $OBJ_OUT sec $SEC_NAME
		fi

		echo "$NATIVE_DEV" > $RUNDIR/device.state
	fi
else
	FILE=$RUNDIR/device.state
	if [ -f $FILE ]; then
		DEV=$(cat $FILE)
		echo "Removed BPF program from device $DEV"
		tc qdisc del dev $DEV clsact 2> /dev/null || true
		rm $FILE
	fi
fi
