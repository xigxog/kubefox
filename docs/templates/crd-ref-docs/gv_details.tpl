{{- define "gvDetails" -}}
{{- $gv := . -}}

## {{ $gv.GroupVersionString }}

{{ $gv.Doc }}

{{ range $gv.SortedTypes }}
{{- if .GVK }}
{{ template "type" . }}
{{ end }}
{{ end }}

## Types

{{ range $gv.SortedTypes }}
{{- if not .GVK }}
{{ template "type" . }}
{{ end }}
{{ end }}

{{- end -}}
