#!/bin/bash
# Copyright 2023 XigXog
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.
#
# SPDX-License-Identifier: MPL-2.0

source "$(dirname "${BASH_SOURCE[0]}")/setup.sh"

function teardown {
    kill $(jobs -p)
    rm -rf /tmp/kubefox

    echo
    echo "Patching Platform to disable debug..."
    kubectl patch \
        -n ${PLATFORM_NS} platform/debug \
        --type merge \
        --patch '{"spec":{"telemetry":{"logs":{"format":"json","level":"info"}},"debug":{"enabled":false,"brokerAddr":""}}}'
    sleep 1
    echo

    echo "Waiting for Platform to be ready...."
    kubectl wait --for=condition=available --timeout 2m -n ${PLATFORM_NS} platform/debug

    exit 0
}

# enable job control
set -m
# kill all jobs on exit
trap 'teardown' SIGINT SIGTERM

INSTANCE_NS="kubefox-system"
PLATFORM_NS="kubefox-debug"
IP=$(hostname -I | awk '{print $1}')

HOST_IP=${HOST_IP:-"${IP}"}

mkdir -p /tmp/kubefox
kubectl get configmap -n ${INSTANCE_NS} kubefox-root-ca -o jsonpath='{.data.ca\.crt}' >/tmp/kubefox/ca.crt
echo "Wrote root certificate to /tmp/kubefox/ca.crt"

kubectl create token -n ${PLATFORM_NS} debug-broker >/tmp/kubefox/broker-token
echo "Wrote broker service account token to /tmp/kubefox/broker-token"

kubectl create token -n ${PLATFORM_NS} debug-httpsrv >/tmp/kubefox/httpsrv-token
echo "Wrote httpsrv service account token to /tmp/kubefox/httpsrv-token"

echo "Port forwarding to Vault..."
kubectl port-forward -n ${INSTANCE_NS} service/kubefox-vault 8200:8200 &
sleep 1
echo

echo "Bootstrapping broker..."
go run ./components/bootstrap/main.go \
    -instance=kubefox \
    -platform-namespace=${PLATFORM_NS} \
    -component=broker \
    -component-service-name=debug-broker.${PLATFORM_NS} \
    -component-ip=${IP} \
    -vault-url=https://127.0.0.1:8200 \
    -log-format=console \
    -log-level=debug \
    -token-path=/tmp/kubefox/broker-token
echo

echo "Patching Platform to enable debug..."
kubectl patch \
    -n ${PLATFORM_NS} platform/debug \
    --type merge \
    --patch '{"spec":{"telemetry":{"logs":{"format":"console","level":"debug"}},"debug":{"enabled":true,"brokerAddr":"'${IP}':6060"}}}'
sleep 1
echo

echo "Waiting for Platform to be ready...."
kubectl wait --for=condition=available --timeout 2m -n ${PLATFORM_NS} platform/debug
echo

echo "Port forwarding to NATS..."
kubectl port-forward -n ${PLATFORM_NS} service/debug-nats 4222:4222 &
sleep 1
echo

echo "Start a local instance of the broker with the following commands or use"
echo "the VSCode launch configurations named 'broker' and 'httpsrv'."
echo
cat <<EOF
# Start broker
go run ./components/broker/ \\
    -instance=kubefox \\
    -platform=debug \\
    -namespace=kubefox-debug \\
    -grpc-addr=0.0.0.0:6060 \\
    -telemetry-addr=false \\
    -health-addr=false \\
    -log-format=console \\
    -log-level=debug \\
    -token-path=/tmp/kubefox/broker-token

# Start httpsrv
go run ./components/httpsrv/ \\
    -platform=debug \\
    -name=httpsrv \\
    -hash=debug \\
    -pod=debug \\
    -https-addr=false \\
    -broker-addr=127.0.0.1:6060 \\
    -health-addr=false \\
    -log-format=console \\
    -log-level=debug \\
    -token-path=/tmp/kubefox/httpsrv-token
EOF
echo

echo "🦊🏁🌟 Debug environment ready! 🌟🏁🦊"
echo

wait
