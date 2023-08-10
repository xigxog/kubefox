#!/bin/sh

# Disable core dumps.
ulimit -c 0

KUBEFOX_PLATFORM="${KUBEFOX_PLATFORM:-kubefox}"
KUBEFOX_PLATFORM_NAMESPACE="${KUBEFOX_PLATFORM_NAMESPACE:-kubefox-system}"
KUBEFOX_ROOT_CA_SECRET="${KUBEFOX_ROOT_CA_SECRET:-$KUBEFOX_PLATFORM-root-ca}"
KUBEFOX_UNSEAL_KEY_SECRET="${KUBEFOX_UNSEAL_KEY_SECRET:-$KUBEFOX_PLATFORM-unseal-key}"

VAULT_TEMP_DIR="${VAULT_TEMP_DIR:-/tmp/vault}"
VAULT_CONF_FILE="${VAULT_CONF_FILE:-$VAULT_TEMP_DIR/conf.hcl}"
VAULT_DATA_DIR="${VAULT_DATA_DIR:-/vault/data}"
VAULT_PLUGIN_DIR="${VAULT_PLUGIN_DIR:-$VAULT_DATA_DIR/plugins}"
VAULT_TLS_DIR="${VAULT_TLS_DIR:-$VAULT_DATA_DIR/tls}"
VAULT_LOG_FORMAT="${VAULT_LOG_FORMAT:-json}"

KUBERNETES_SERVICE_HOST="${KUBERNETES_SERVICE_HOST:-kubernetes.default}"
KUBERNETES_SERVICE_PORT="${KUBERNETES_SERVICE_PORT:-443}"

NATS_URL="${NATS_URL:-nats://nats.$KUBEFOX_PLATFORM_NAMESPACE:4222}"
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
        // manage NATS plugin
        path "sys/plugins/reload/backend" {
            capabilities = ["create", "update", "sudo"]
        }
        path "sys/plugins/catalog/secret/nats" {
            capabilities = ["create", "update", "sudo"]
        }
        path "sys/mounts/nats" {
            capabilities = ["read", "update"]
        }
        path "sys/mounts/nats/tune" {
            capabilities = ["update"]
        }
        path "nats/config" {
            capabilities = ["create", "update"]
        }
    ' | vault policy write nats-plugin-manager -
    echo '
        // manage KV stores
        path "sys/mounts/kubefox/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
        path "kubefox/*" {
            capabilities = ["create", "read", "update", "patch", "delete", "list"]
        }
    ' | vault policy write kubefox-kv-manager -
    echo '
        // issue operator certs
        path "pki_int/issue/operator" {
            capabilities = ["create", "update"]
        }
        // issue broker certs
        path "pki_int/issue/broker" {
            capabilities = ["create", "update"]
        }
    ' | vault policy write kubefox-cert-issuer -
    echo '
        path "nats/jwt/*" {
            capabilities = ["create"]
        }
        path "kubefox/*" {
            capabilities = ["read"]
        }
    ' | vault policy write kubefox-kv-reader -

    vault write auth/kubernetes/role/kubefox-vault \
        bound_service_account_names=$KUBEFOX_PLATFORM-vault \
        bound_service_account_namespaces=$KUBEFOX_PLATFORM_NAMESPACE \
        token_policies=nats-plugin-manager
    vault write auth/kubernetes/role/kubefox-operator \
        bound_service_account_names=$KUBEFOX_PLATFORM-operator \
        bound_service_account_namespaces=$KUBEFOX_PLATFORM_NAMESPACE \
        token_policies=kubefox-kv-manager,kubefox-cert-issuer
    vault write auth/kubernetes/role/kubefox-api-server \
        bound_service_account_names=$KUBEFOX_PLATFORM-api-server \
        bound_service_account_namespaces=$KUBEFOX_PLATFORM_NAMESPACE \
        token_policies=kubefox-kv-manager
    vault write auth/kubernetes/role/kubefox-broker \
        bound_service_account_names=$KUBEFOX_PLATFORM-broker \
        bound_service_account_namespaces=$KUBEFOX_PLATFORM_NAMESPACE \
        token_policies=kubefox-kv-reader

    echo "Policies and roles for Kubernetes auth applied."

    hundred_yrs="876000h"
    ten_yrs="87600h"

    vault secrets enable pki
    vault secrets tune -max-lease-ttl="$hundred_yrs" pki
    vault write pki/config/urls \
        issuing_certificates="https://$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE:8200/v1/pki/ca" \
        crl_distribution_points="https://$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE:8200/v1/pki/crl" \
        >/dev/null
    vault write -field=certificate pki/root/generate/internal \
        common_name="KubeFox ($KUBEFOX_PLATFORM) Root CA" \
        issuer_name="$KUBEFOX_PLATFORM-root" \
        ttl="$hundred_yrs" |
        sed 's/\\n/\n/g' | sed 's/"//g' \
        >$VAULT_TLS_DIR/ca.crt
    cat >"$VAULT_TEMP_DIR/ca.yaml" <<EOF
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: $KUBEFOX_ROOT_CA_SECRET
data:
  ca.crt: $(cat "$VAULT_TLS_DIR/ca.crt" | base64 -w 0)
EOF
    kubectl apply -f "$VAULT_TEMP_DIR/ca.yaml"
    rm -f "$VAULT_TEMP_DIR/ca.yaml"

    vault write pki/roles/global allow_any_name=true >/dev/null
    echo "Root CA created, cert written to Kubernetes Secret $KUBEFOX_ROOT_CA_SECRET."

    vault secrets enable -path=pki_int pki
    vault secrets tune -max-lease-ttl="$hundred_yrs" pki_int
    vault write pki/config/urls \
        issuing_certificates="https://$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE:8200/v1/pki_int/ca" \
        crl_distribution_points="https://$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE:8200/v1/pki_int/crl" \
        >/dev/null
    vault pki issue \
        -issuer_name=$KUBEFOX_PLATFORM-intermediate \
        /pki/issuer/$KUBEFOX_PLATFORM-root \
        /pki_int/ \
        common_name="KubeFox ($KUBEFOX_PLATFORM) Intermediate CA" \
        ttl="$hundred_yrs" \
        >/dev/null

    # Role is used to generate certs of Vault's HTTPS listener.
    vault write pki_int/roles/vault \
        issuer_ref="$KUBEFOX_PLATFORM-intermediate" \
        allow_localhost=true \
        allowed_domains="$KUBEFOX_PLATFORM-vault,$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE" \
        allow_bare_domains=true \
        max_ttl="$ten_yrs" \
        >/dev/null
    # Role is used by operator to generate certs for it's gRPC server and HTTPS
    # server.
    vault write pki_int/roles/operator \
        issuer_ref="$KUBEFOX_PLATFORM-intermediate" \
        allow_localhost=true \
        allowed_domains="$KUBEFOX_PLATFORM-operator,$KUBEFOX_PLATFORM-operator.$KUBEFOX_PLATFORM_NAMESPACE" \
        allow_bare_domains=true \
        max_ttl="$ten_yrs" \
        >/dev/null
    # Role is used by operator to generate certs for broker's gRPC servers.
    vault write pki_int/roles/broker \
        issuer_ref="$KUBEFOX_PLATFORM-intermediate" \
        allow_localhost=true \
        allowed_domains="localhost" \
        allow_bare_domains=true \
        max_ttl="$ten_yrs" \
        >/dev/null
    echo "Intermediate CA and roles created."

    # Issue cert for Vault's HTTPS listener.
    vault write pki_int/issue/vault \
        common_name="$KUBEFOX_PLATFORM-vault.$KUBEFOX_PLATFORM_NAMESPACE" \
        alt_names="$KUBEFOX_PLATFORM-vault,localhost" \
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

cur_ver=$(vault read sys/mounts/nats 2>/dev/null | jq -r '.data.plugin_version')
# If there is no current version then the plugin has not been enabled yet.
if [ -z $cur_ver ]; then
    vault secrets enable -plugin-version=$ver nats
    vault write nats/config \
        service_url=$NATS_URL \
        >/dev/null
fi
# Always set plugin version to be same as version present on the container.
vault secrets tune -plugin-version=$ver nats
# This is not technically needed as we restart Vault below but ensures plugin
# will load properly.
vault plugin reload -plugin nats
echo "NATS plugin version $ver enabled."

vault token revoke -self
# Kill Vault using Unix socket listener so we can restart it.
kill $vault_pid
echo "Vault initialized."

# Restart Vault using HTTPS listener.
export VAULT_ADDR="https://127.0.0.1:8200"

cat >"$VAULT_CONF_FILE" <<EOF
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
