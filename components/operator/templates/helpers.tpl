{{- define "labels" -}}
{{ include "selectors" . }}
app.kubernetes.io/version: {{ .Component.GitRef | default .Component.Commit }}
{{- with .Component.GitRef }}
kubefox.xigxog.io/component-git-ref: {{ . }}
{{- end }}
{{- with .App.Commit }}
kubefox.xigxog.io/app-commit: {{ . }}
{{- end }}
{{- with .App.GitRef }}
kubefox.xigxog.io/app-git-ref: {{ . }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Instance.Name }}-operator
{{ .Labels | toYaml }}
{{- end }}

{{- define "selectors" -}}
app.kubernetes.io/instance: {{ .Instance.Name }}
{{- with .Platform.Name }}
kubefox.xigxog.io/platform: {{ . }}
{{- end }}
{{- with .App.Name }}
app.kubernetes.io/name: {{ . }} 
{{- end }}
app.kubernetes.io/component: {{ .Component.Name }}
kubefox.xigxog.io/component: {{ .Component.Name }}
kubefox.xigxog.io/component-commit: {{ .Component.Commit }}
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

{{- define "podSecurityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 100
  runAsGroup: 1000
  fsGroup: 1000
  fsGroupChangePolicy: OnRootMismatch
{{- end }}

{{- define "containerSecurityContext" -}}
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
{{- end }}