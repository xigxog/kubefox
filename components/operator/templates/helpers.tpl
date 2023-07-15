{{- define "fullname" -}}
{{ printf "%s-%s" .Platform.Name .Component.Name }}
{{- end }}

{{- define "labels" -}}
{{ include "selectors" . }}
app.kubernetes.io/part-of: {{ .System.Name }}
app.kubernetes.io/managed-by: {{ .Platform.Name }}-operator
k8s.kubefox.io/component: {{ .Component.Name }}
{{- with .Component.GitHash }}
app.kubernetes.io/version: {{ . }}
k8s.kubefox.io/component-git-hash: {{ . }}
{{- end }}
{{- with .System.Name }}
k8s.kubefox.io/system: {{ . }}
{{- end }}
{{- with .System.Id }}
k8s.kubefox.io/system-id: {{ . }}
{{- end }}
{{- with .System.Ref }}
k8s.kubefox.io/system-ref: {{ . }}
{{- end }}
{{- with .System.GitHash }}
k8s.kubefox.io/system-git-hash: {{ . }}
{{- end }}
{{- with .System.GitRef }}
k8s.kubefox.io/system-git-ref: {{ . }}
{{- end }}
{{- with .Environment.Name }}
k8s.kubefox.io/environment: {{ . }}
{{- end }}
{{- with .Environment.Id }}
k8s.kubefox.io/environment-id: {{ . }}
{{- end }}
{{- with .Environment.Ref }}
k8s.kubefox.io/environment-ref: {{ . }}
{{- end }}
{{- with .Config.Name }}
k8s.kubefox.io/config: {{ . }}
{{- end }}
{{- with .Config.Id }}
k8s.kubefox.io/config-id: {{ . }}
{{- end }}
{{- with .Config.Ref }}
k8s.kubefox.io/config-ref: {{ . }}
{{- end }}
{{ .Labels | toYaml }}
{{- end }}

{{- define "selectors" -}}
app.kubernetes.io/name: {{ .Component.Name }}
app.kubernetes.io/instance: {{ .Component.Name }}-{{ .Platform.Name }}
{{- end }}

{{- define "metadata" -}}
metadata:
  name: {{ include "fullname" . }}
  namespace: {{ .System.Namespace }}
  labels:
    {{- include "labels" . | nindent 4 }}
  {{- with .Owner }}
  ownerReferences:
    - {{- . | toYaml | nindent 6 }}
  {{- end }}
{{- end }}

{{- define "broker" -}}
name: broker
image: {{ .Broker.Image | default (printf "ghcr.io/xigxog/kubefox/broker:%v" .Platform.Version) }}
imagePullPolicy: {{ .Broker.ImagePullPolicy | default "IfNotPresent" }}
args:
  - {{ .Broker.Type | default "component" }}
  - --platform={{ .Platform.Name }}
  - --system={{ .System.Name }}
  - --component={{ .Component.Name }}
  {{- with .Component.GitHash }}
  - --component-hash={{ . }}
  {{- end }}
  - --operator-addr={{ printf "%s-operator.%s:7070" .Platform.Name .Platform.Namespace }}
  - --nats-addr={{ printf "%s-nats.%s:4222" .Platform.Name .Platform.Namespace }}
  - --namespace={{ .Platform.Namespace }}
  - --telemetry-agent-addr=$(HOST_IP):4318
  - --health-addr=0.0.0.0:1111
  {{- if .DevMode }}
  - --dev
  {{- end }}
env:
  - name: HOST_IP
    valueFrom:
      fieldRef:
        apiVersion: v1
        fieldPath: status.hostIP
ports:
  - name: health
    containerPort: 1111
    protocol: TCP
livenessProbe:
  httpGet:
    port: health
{{- end }}

{{- define "roleBinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
{{ include "metadata" . }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ include "fullname" . }}
subjects:
  - kind: ServiceAccount
    name: {{ include "fullname" . }}
    namespace: {{ .System.Namespace }}
{{- end }}

{{- define "clusterRoleBinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
{{ include "metadata" . }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ include "fullname" . }}
subjects:
  - kind: ServiceAccount
    name: {{ include "fullname" . }}
    namespace: {{ .System.Namespace }}
{{- end }}

{{- define "serviceAccount" -}}
apiVersion: v1
kind: ServiceAccount
{{ include "metadata" . }}
{{- end }}