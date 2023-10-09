#!/bin/sh

# Debug
# set -x

KUBEFOX_HOME="${KUBEFOX_HOME:-/tmp/kubefox}"
KUBEFOX_INSTANCE="${KUBEFOX_INSTANCE:-kubefox}"
KUBEFOX_INSTANCE_NAMESPACE="${KUBEFOX_INSTANCE_NAMESPACE:-kubefox}"
KUBEFOX_PLATFORM="${KUBEFOX_PLATFORM:-kubefox}"
KUBEFOX_PLATFORM_NAMESPACE="${KUBEFOX_PLATFORM_NAMESPACE:-kubefox}"
KUBEFOX_PLATFORM_VAULT_NAME="${KUBEFOX_PLATFORM_VAULT_NAME:-missing}"
KUBEFOX_COMPONENT="${KUBEFOX_COMPONENT:-kubefox}"
KUBEFOX_COMPONENT_IP="${KUBEFOX_COMPONENT_IP:-127.0.0.1}"

VAULT_ROLE="${VAULT_ROLE:-$KUBEFOX_PLATFORM_VAULT_NAME-$KUBEFOX_COMPONENT}"
export VAULT_CACERT="$KUBEFOX_HOME/ca.crt"
export VAULT_ADDR="https://$KUBEFOX_INSTANCE-vault.$KUBEFOX_INSTANCE_NAMESPACE:8200"
export VAULT_FORMAT="json"

ten_yrs="87600h"

mkdir -p "$KUBEFOX_HOME"

# Login to Vault using Kubernetes auth.
jwt=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
token=$(vault write /auth/kubernetes/login jwt="$jwt" role="$VAULT_ROLE" | jq -r '.auth.client_token')
vault login "$token" >/dev/null

# Issue TLS cert
cname="$KUBEFOX_PLATFORM-$KUBEFOX_COMPONENT.$KUBEFOX_PLATFORM_NAMESPACE"
vault write pki/int/platform/$KUBEFOX_PLATFORM_VAULT_NAME/issue/$KUBEFOX_COMPONENT \
    common_name="$cname" \
    alt_names="$KUBEFOX_COMPONENT@$cname,localhost" \
    ip_sans="$KUBEFOX_COMPONENT_IP,127.0.0.1" \
    ttl="$ten_yrs" \
    >"$KUBEFOX_HOME/cert.json"
jq -r '.data.certificate, .data.issuing_ca' "$KUBEFOX_HOME/cert.json" >"$KUBEFOX_HOME/tls.crt"
jq -r '.data.private_key' "$KUBEFOX_HOME/cert.json" >"$KUBEFOX_HOME/tls.key"
