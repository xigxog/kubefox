{{- define "labels" -}}
{{ include "selectors" . }}
app.kubernetes.io/managed-by: {{ printf "%s-operator" .Instance.Name | quote }}
kubefox.xigxog.io/runtime-version: {{ .BuildInfo.Version | quote }}
{{- with .Component.Labels }}
{{ . | toYaml }}
{{- end }}
{{- end }}

{{- define "annotations" -}}
{{- with .Hash }}
kubefox.xigxog.io/template-data-hash: {{ . | quote }}
{{- end }}
{{- with .Component.Annotations }}
{{ . | toYaml }}
{{- end }}
{{- end }}

{{- define "selectors" -}}
app.kubernetes.io/instance: {{ .Instance.Name | quote }}
{{- with .Platform.Name }}
kubefox.xigxog.io/platform: {{ . | quote }}
{{- end }}
{{- with .Component.App }}
app.kubernetes.io/name: {{ . | quote }} 
{{- end }}
{{- with .Component.Name }}
app.kubernetes.io/component: {{ . | quote }}
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
  annotations:
    {{- include "annotations" . | nindent 4 }}
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
{{- with .Values.GOMEMLIMIT }}
- name: GOMEMLIMIT
  value: {{ . | quote }}
{{- end }}
{{- with .Values.GOMAXPROCS }}
- name: GOMAXPROCS
  value: {{ . | quote }}
{{- end }}
{{- end }}

{{- define "podSpec" -}}
serviceAccountName: {{ componentFullName }}
securityContext:
  runAsNonRoot: true
  runAsUser: 100
  runAsGroup: 1000
  fsGroup: 1000
  fsGroupChangePolicy: OnRootMismatch

{{- with .Component.ImagePullSecret }}
imagePullSecrets:
  - name: {{ . }}
{{- end }}

{{- with .Component.NodeSelector }}
nodeSelector:
  {{- . | toYaml | nindent 2 }}
{{- end }}

{{- with .Component.NodeName }}
nodeName: {{ . | quote  }}
{{- end }}

{{- with .Component.Affinity }}
affinity:
  {{- . | toYaml | nindent 2 }}
{{- end }}

{{- with .Component.Tolerations }}
tolerations:
  {{- . | toYaml | nindent 2 }}
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
  - -platform-vault-name={{ platformVaultName }}
  - -component-vault-name={{ componentVaultName }}
  - -component-service-name={{ printf "%s.%s" componentFullName namespace }}
  - -component-ip=$(KUBEFOX_COMPONENT_IP)
  - -vault-url={{ .Values.vaultURL }}
  - -log-format={{ .Logger.Format | default "json" }}
  - -log-level={{ .Logger.Level | default "info" }}
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