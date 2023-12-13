{{- define "type" -}}
{{- $type := . -}}
{{- if and (markdownShouldRenderType $type) (not $type.IsAlias) -}}

### {{ $type.Name }}

{{ $type.Doc }}

{{ if $type.References -}}
<p style="font-size:.6rem;">
Used by:<br>
{{ range $type.SortedReferences }}
- <a href=#{{ .Name | lower }}>{{ .Name }}</a><br>
{{- end }}
</p>
{{- end }}

{{ if $type.Members -}}
| Field | Type | Description | Validation |
| ----- | ---- | ----------- | ---------- |
{{ if $type.GVK -}}
| `apiVersion` | string | `{{ $type.GVK.Group }}/{{ $type.GVK.Version }}` | |
| `kind` | string | `{{ $type.GVK.Kind }}` | |
{{ end -}}

{{ range $type.Members -}}
| `{{ .Name  }}` | <div style="white-space:nowrap">{{ markdownRenderType . }}<div> | <div style="max-width:30rem">{{ template "type_members" . }}</div> | <div style="white-space:nowrap">{{ markdownRenderValidation . }}</div> |
{{ end -}}
{{ end -}}

{{- end }}
{{- end }}