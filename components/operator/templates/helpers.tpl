{{- define "labels" -}}
{{ include "selectors" . }}
{{- with .Component.Name }}
app.kubernetes.io/component: {{ . | quote }}
{{- end }}
{{- with .Component.Commit }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
{{- with .App.Commit }}
kubefox.xigxog.io/app-commit: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ printf "%s-operator" .Instance.Name | quote }}
{{ .Labels | toYaml }}
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

{{- define "containerSecurityContext" -}}
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
{{- end }}

{{- define "env" -}}
{{- with .Component.Name }}
- name: KUBEFOX_COMPONENT
  value: {{ . }}
{{- end }}
{{- with .Component.Commit }}
- name: KUBEFOX_COMMIT
  value: {{ . }}
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

{{- if .Component.Tolerations }}
tolerations:
  {{- .Component.Tolerations | toYaml | nindent 2 }}
{{- else if .App.Tolerations }}
tolerations:
  {{- .App.Tolerations | toYaml | nindent 2 }}
{{- end }}

{{- if .Component.Affinity }}
affinity:
  {{- .Component.Affinity | toYaml | nindent 2 }}
{{- else if .App.Affinity }}
affinity:
  {{- .App.Affinity | toYaml | nindent 2 }}
{{- end }}
{{- end }}