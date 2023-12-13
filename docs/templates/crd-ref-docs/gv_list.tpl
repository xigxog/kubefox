{{- define "gvList" -}}
{{- $groupVersions := . -}}

# Kubernetes CRDs

{{- range $groupVersions }}
{{ template "gvDetails" . }}
{{- end }}

{{- end -}}
