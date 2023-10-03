#!/bin/sh

# Debug
# set -x

# Disable core dumps.
ulimit -c 0

KUBEFOX_INSTANCE="${KUBEFOX_INSTANCE:-main}"
KUBEFOX_NAMESPACE="${KUBEFOX_NAMESPACE:-kubefox-system}"
KUBEFOX_ROOT_CA_CM="${KUBEFOX_ROOT_CA_CM:-$KUBEFOX_INSTANCE-root-ca}"
KUBEFOX_UNSEAL_KEY_SECRET="${KUBEFOX_UNSEAL_KEY_SECRET:-$KUBEFOX_INSTANCE-unseal-key}"

VAULT_TEMP_DIR="${VAULT_TEMP_DIR:-/tmp/vault}"
VAULT_CONF_FILE="${VAULT_CONF_FILE:-$VAULT_TEMP_DIR/conf.hcl}"
VAULT_DATA_DIR="${VAULT_DATA_DIR:-/vault/data}"
VAULT_PLUGIN_DIR="${VAULT_PLUGIN_DIR:-$VAULT_DATA_DIR/plugins}"
VAULT_TLS_DIR="${VAULT_TLS_DIR:-$VAULT_DATA_DIR/tls}"
VAULT_LOG_FORMAT="${VAULT_LOG_FORMAT:-json}"

KUBERNETES_SERVICE_HOST="${KUBERNETES_SERVICE_HOST:-kubernetes.default}"
KUBERNETES_SERVICE_PORT="${KUBERNETES_SERVICE_PORT:-443}"

CONTAINER_PLUGIN_DIR="${CONTAINER_PLUGIN_DIR:-/xigxog/vault/plugins}"

export VAULT_ADDR="unix://$VAULT_TEMP_DIR/vault.sock"
export VAULT_FORMAT="json"

wait_for_vault() {
    code=$(
        vault status >/dev/null 2>&1
        echo $?
    )
    while [ $code -eq 1 ]; do
        echo "Waiting for Vault to be ready..."
        sleep 1
        code=$(
            vault status >/dev/null 2>&1
            echo $?
        )
    done
    echo "Vault is ready."
}

unseal() {
    sealed=$(vault status | jq -r '.sealed')
    if $sealed; then
        echo "Vault is currently sealed."
        key=$(kubectl get secret "$KUBEFOX_UNSEAL_KEY_SECRET" -o jsonpath='{.data.key}' | base64 -d)
        vault operator unseal "$key" >/dev/null
        unset key
    fi
    echo "Vault unsealed."
}

mkdir -p "$VAULT_TEMP_DIR" "$VAULT_DATA_DIR" "$VAULT_PLUGIN_DIR" "$VAULT_TLS_DIR"

# Start Vault using Unix socket to securely initialize it. Once initialized
# Vault is restarted using an HTTPS listener.
echo "Starting Vault using Unix socket listener for secure initialization..."

cat >"$VAULT_CONF_FILE" <<EOF
disable_mlock = true
disable_clustering = true
api_addr = "$VAULT_ADDR"
plugin_directory = "$VAULT_PLUGIN_DIR"
log_format = "standard"

storage "file" {
    path = "$VAULT_DATA_DIR"
}

listener "unix" {
    address = "$VAULT_TEMP_DIR/vault.sock"
}
EOF

vault server -config="$VAULT_CONF_FILE" >"$VAULT_TEMP_DIR/init.log" 2>&1 &
vault_pid=$!
wait_for_vault

# From here on if any command fails we want to exit script.
set -e

init_file="$VAULT_TEMP_DIR/init.json"
# Ensures cleanup if script exits early.
trap 'rm -f "$init_file"; kill $vault_pid' EXIT HUP TERM INT

init=$(vault status | jq -r '.initialized')
if ! "$init"; then
    echo "Initializing Vault..."

    vault operator init \
        -key-shares=1 \
        -key-threshold=1 \
        >"$init_file"

    key=$(jq -r '.unseal_keys_b64[0]' "$init_file")
    cat >"$VAULT_TEMP_DIR/key.yaml" <<EOF
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: $KUBEFOX_UNSEAL_KEY_SECRET
data:
  key: $(echo -n "$key" | base64 -w 0)
EOF
    kubectl apply -f "$VAULT_TEMP_DIR/key.yaml"
    rm -f "$VAULT_TEMP_DIR/key.yaml"
    echo "Unseal key stored in Kubernetes Secret $KUBEFOX_UNSEAL_KEY_SECRET."

    vault operator unseal "$key" >/dev/null
    echo "Vault unsealed."

    token=$(jq -r '.root_token' "$init_file")
    vault login "$token" >/dev/null
    unset token
    rm -f "$init_file"

    vault auth enable kubernetes
    vault write auth/kubernetes/config \
        kubernetes_host="https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT"
    echo "Kubernetes auth enabled."

    echo '
        // administer NATS plugin
        path "sys/plugins/catalog/secret/nats" {
            capabilities = ["create", "update", "sudo"]
        }
        path "sys/plugins/reload/backend" {
            capabilities = ["create", "update", "sudo"]
        }
        path "sys/mounts" {
            capabilities = ["list"]
        }
        path "sys/mounts/nats/*/tune" {
            capabilities = ["update"]
        }
    ' | vault policy write plugin-nats-admin -
    echo '
        // manage Platform policies and K8s auth roles
        path "sys/policies/acl/kubefox-*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
        path "auth/kubernetes/role/kubefox-*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
    ' | vault policy write auth-kubernetes-manager -
    echo '
        // manage NATS mounts
        path "sys/mounts/nats/platform/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
        path "nats/platform/*" {
            capabilities = ["create", "update"]
        }
    ' | vault policy write nats-manager -
    echo '
        // manage Platform intermediate PKI mounts
        path "sys/mounts/pki/int/platform/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
        path "pki/root/root/sign-intermediate" {
            capabilities = ["create", "update"]
        }
        path "pki/int/platform/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
    ' | vault policy write pki-int-manager -
    echo '
        // manage KubeFox KV mount
        path "sys/mounts/kubefox/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
        path "kubefox/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
    ' | vault policy write kv-kubefox-manager -
    echo '
        // read KubeFox KV store
        path "kubefox/*" {
            capabilities = ["read", "list"]
        }
    ' | vault policy write kv-kubefox-reader -

    vault write auth/kubernetes/role/kubefox-vault \
        bound_service_account_names=$KUBEFOX_INSTANCE-vault \
        bound_service_account_namespaces=$KUBEFOX_NAMESPACE \
        token_policies=plugin-nats-admin
    vault write auth/kubernetes/role/kubefox-operator \
        bound_service_account_names=$KUBEFOX_INSTANCE-operator \
        bound_service_account_namespaces=$KUBEFOX_NAMESPACE \
        token_policies=nats-manager,auth-kubernetes-manager,pki-int-manager,kv-kubefox-manager

    echo "Cluster policies and roles for Kubernetes auth applied."

    hundred_yrs="876000h"
    ten_yrs="87600h"

    vault secrets enable -path=pki/root pki
    vault secrets tune -max-lease-ttl="$hundred_yrs" pki/root/
    vault write pki/root/config/urls \
        issuing_certificates="https://$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE:8200/v1/pki/root/ca" \
        crl_distribution_points="https://$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE:8200/v1/pki/root/crl" \
        >/dev/null
    vault write pki/root/roles/global allow_any_name=true >/dev/null
    vault write -field=certificate pki/root/root/generate/internal \
        common_name="KubeFox ($KUBEFOX_INSTANCE) Root CA" \
        issuer_name="$KUBEFOX_INSTANCE-root" \
        ttl="$hundred_yrs" |
        sed 's/\\n/\n/g' | sed 's/"//g' \
        >"$VAULT_TLS_DIR/ca.crt"
    kubectl delete configmap $KUBEFOX_ROOT_CA_CM 2>/dev/null || true
    kubectl create configmap $KUBEFOX_ROOT_CA_CM --from-file "$VAULT_TLS_DIR/ca.crt"
    echo "Root CA created, cert written to Kubernetes Secret $KUBEFOX_ROOT_CA_CM."

    vault secrets enable -path=pki/int/instance/ pki
    vault secrets tune -max-lease-ttl="$hundred_yrs" pki/int/instance
    vault write pki/int/instance/config/urls \
        issuing_certificates="https://$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE:8200/v1/pki/int/instance/ca" \
        crl_distribution_points="https://$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE:8200/v1/pki/int/instance/crl" \
        >/dev/null
    vault pki issue \
        -issuer_name=$KUBEFOX_INSTANCE-intermediate \
        /pki/root/issuer/$KUBEFOX_INSTANCE-root \
        /pki/int/instance/ \
        common_name="KubeFox ($KUBEFOX_INSTANCE) Intermediate CA" \
        ttl="$hundred_yrs" \
        >/dev/null

    # Role is used to generate certs of Vault's HTTPS listener.
    vault write pki/int/instance/roles/vault \
        issuer_ref="$KUBEFOX_INSTANCE-intermediate" \
        allow_localhost=true \
        allowed_domains="$KUBEFOX_INSTANCE-vault,$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE" \
        allow_bare_domains=true \
        max_ttl="$ten_yrs" \
        >/dev/null
    echo "Intermediate CA and roles created."

    # Issue cert for Vault's HTTPS listener.
    vault write pki/int/instance/issue/vault \
        common_name="$KUBEFOX_INSTANCE-vault.$KUBEFOX_NAMESPACE" \
        alt_names="$KUBEFOX_INSTANCE-vault,localhost" \
        ip_sans="127.0.0.1" \
        ttl="$ten_yrs" \
        >"$VAULT_TEMP_DIR/cert.json"
    jq -r '.data.certificate, .data.issuing_ca' "$VAULT_TEMP_DIR/cert.json" >"$VAULT_TLS_DIR/tls.crt"
    jq -r '.data.private_key' "$VAULT_TEMP_DIR/cert.json" >"$VAULT_TLS_DIR/tls.key"
    rm -f "$VAULT_TEMP_DIR/cert.json"
    echo "Vault HTTPS listener certs generated."

    # Ensure root token cannot be used again. If a new root token is needed the
    # unseal key can be used to generate it.
    vault token revoke -self
fi

# This needs to be done after init otherwise server using Unix sockets will fail
# to start as the CA might not exist yet.
export VAULT_CACERT="$VAULT_TLS_DIR/ca.crt"
unseal

# Login using Kubernetes auth.
jwt=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
token=$(vault write /auth/kubernetes/login jwt="$jwt" role=kubefox-vault | jq -r '.auth.client_token')
vault login "$token" >/dev/null

# Setup the NATS plugin present on the container.
ver=$(cat "$CONTAINER_PLUGIN_DIR/nats-plugin.version")
sha=$(cat "$CONTAINER_PLUGIN_DIR/nats-plugin-$ver.sha256sum" | cut -d" " -f1)
cp "$CONTAINER_PLUGIN_DIR/nats-plugin-$ver" "$VAULT_PLUGIN_DIR"
chmod 700 "$VAULT_PLUGIN_DIR/nats-plugin-$ver"
# This is safe to run even if plugin has already been registered.
vault plugin register \
    -sha256=$sha \
    -version=$ver \
    -command=nats-plugin-$ver \
    secret nats

### TODO loop over all existing enabled nats secret engines and update them
# cur_ver=$(vault read sys/mounts/nats 2>/dev/null | jq -r '.data.plugin_version')
# # If there is no current version then the plugin has not been enabled yet.
# if [ -z $cur_ver ]; then
#     vault secrets enable -plugin-version=$ver nats
#     vault write nats/config \
#         service_url=$NATS_URL \
#         >/dev/null
# fi
# # Always set plugin version to be same as version present on the container.
# vault secrets tune -plugin-version=$ver nats
###

# # This is not technically needed as we restart Vault below but ensures plugin
# # will load properly.
# vault plugin reload -plugin nats
echo "NATS plugin version $ver enabled."

vault token revoke -self
# Kill Vault using Unix socket listener so we can restart it.
kill $vault_pid
echo "Vault initialized."

# Restart Vault using HTTPS listener.
export VAULT_ADDR="https://127.0.0.1:8200"

cat >"$VAULT_CONF_FILE" <<EOF
ui = true

disable_mlock = true
disable_clustering = true
api_addr = "$VAULT_ADDR"
plugin_directory = "$VAULT_PLUGIN_DIR"
log_format = "$VAULT_LOG_FORMAT"

storage "file" {
    path = "$VAULT_DATA_DIR"
}

listener "tcp" {
    tls_disable = 0
    address = "0.0.0.0:8200"
    tls_cert_file = "$VAULT_TLS_DIR/tls.crt"
    tls_key_file  = "$VAULT_TLS_DIR/tls.key"
}
EOF

# Auto unseal using Kubernetes secret.
{
    wait_for_vault
    unseal
} &

exec vault server -config="$VAULT_CONF_FILE"
