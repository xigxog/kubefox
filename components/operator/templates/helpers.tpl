{{- define "labels" -}}
{{ include "selectors" . }}
{{- with .Component.Name }}
app.kubernetes.io/component: {{ . | quote }}
{{- end }}
{{- with .Component.Commit }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ printf "%s-operator" .Instance.Name | quote }}
kubefox.xigxog.io/version: {{ .Instance.Version | quote }}
{{ .ExtraLabels | toYaml }}
{{- end }}

{{- define "selectors" -}}
app.kubernetes.io/instance: {{ .Instance.Name | quote }}
{{- with .Platform.Name }}
kubefox.xigxog.io/platform: {{ . | quote }}
{{- end }}
{{- with .App.Name }}
app.kubernetes.io/name: {{ . | quote }} 
{{- end }}
{{- with .Component.Name }}
kubefox.xigxog.io/component: {{ . | quote }}
{{- end }}
{{- with .Component.Commit }}
kubefox.xigxog.io/component-commit: {{ . | quote }}
{{- end }}
{{- end }}

{{- define "metadata" -}}
metadata:
  name: {{ componentFullName }}
  namespace: {{ namespace }}
  labels:
    {{- include "labels" . | nindent 4 }}
  {{- with .Owner }}
  ownerReferences:
    {{- . | toYaml | nindent 4 }}
  {{- end }}
{{- end }}

{{- define "roleBinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
{{ include "metadata" . }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ componentFullName }}
subjects:
  - kind: ServiceAccount
    name: {{ componentFullName }}
    namespace: {{ namespace }}
{{- end }}

{{- define "clusterRoleBinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
{{ include "metadata" . }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ componentFullName }}
subjects:
  - kind: ServiceAccount
    name: {{ componentFullName }}
    namespace: {{ namespace }}
{{- end }}

{{- define "serviceAccount" -}}
apiVersion: v1
kind: ServiceAccount
{{ include "metadata" . }}
{{- end }}

{{- define "env" -}}
{{- with .Component.Name }}
- name: KUBEFOX_COMPONENT
  value: {{ . | quote }}
{{- end }}
{{- with .Component.Commit }}
- name: KUBEFOX_COMMIT
  value: {{ . | quote }}
{{- end }}
- name: KUBEFOX_HOST_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: KUBEFOX_NODE
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
- name: KUBEFOX_POD
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KUBEFOX_POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
{{- end }}

{{- define "podSpec" -}}
serviceAccountName: {{ componentFullName }}
securityContext:
  runAsNonRoot: true
  runAsUser: 100
  runAsGroup: 1000
  fsGroup: 1000
  fsGroupChangePolicy: OnRootMismatch

{{- with .App.ImagePullSecret }}
imagePullSecrets:
  - name: {{ . }}
{{- end }}

{{- if .Component.NodeSelector }}
nodeSelector:
  {{- .Component.NodeSelector | toYaml | nindent 2 }}
{{- else if .App.NodeSelector }}
nodeSelector:
  {{- .App.NodeSelector | toYaml | nindent 2 }}
{{- end }}

{{- if .Component.NodeName }}
nodeName: {{ .Component.NodeName | quote }}
{{- else if .App.NodeName }}
nodeName: {{ .App.NodeName | quote  }}
{{- end }}

{{- if .Component.Affinity }}
affinity:
  {{- .Component.Affinity | toYaml | nindent 2 }}
{{- else if .App.Affinity }}
affinity:
  {{- .App.Affinity | toYaml | nindent 2 }}
{{- end }}

{{- if .Component.Tolerations }}
tolerations:
  {{- .Component.Tolerations | toYaml | nindent 2 }}
{{- else if .App.Tolerations }}
tolerations:
  {{- .App.Tolerations | toYaml | nindent 2 }}
{{- end }}
{{- end }}

{{- define "securityContext" -}}
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
{{- end }}

{{- define "resources" -}}
{{- with .Component.Resources }}
resources:
  {{- . | toYaml | nindent 2 }}
{{- end }}
{{- end }}

{{- define "probes" -}}
{{- with .Component.LivenessProbe }}
livenessProbe:
  {{- . | toYaml | nindent 2 }}
{{- end }}
{{- with .Component.ReadinessProbe }}
readinessProbe:
  {{- . | toYaml | nindent 2 }}
{{- end }}
{{- with .Component.StartupProbe }}
startupProbe:
  {{- . | toYaml | nindent 2 }}
{{- end }}
{{- end }}

{{- define "bootstrap" -}}
name: bootstrap
image: {{ .Instance.BootstrapImage }}
imagePullPolicy: {{ .Component.ImagePullPolicy | default "IfNotPresent" }}
{{ include "securityContext" . }}
args:
  - -instance={{ .Instance.Name }}
  - -platform={{ .Platform.Name }}
  - -component={{ .Component.Name }}
  - -component-ip=$(KUBEFOX_COMPONENT_IP)
  - -namespace={{ namespace }}
  - -vault-name={{ platformVaultName }}
  - -vault-role={{ printf "%s-%s" platformVaultName .Component.Name }}
  - -vault-addr={{ printf "%s-vault.%s:8200" .Instance.Name .Instance.Namespace }}
  - -log-format={{ logFormat }}
  - -log-level={{ logLevel }}
env:
{{- include "env" . | nindent 2 }}
  - name: KUBEFOX_COMPONENT_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP
envFrom:
  - configMapRef:
      name: {{ name }}-env
volumeMounts:
  - name: root-ca
    mountPath: {{ homePath }}/ca.crt
    subPath: ca.crt
  - name: kubefox
    mountPath: {{ homePath }}
{{- end }}