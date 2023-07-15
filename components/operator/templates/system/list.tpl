apiVersion: v1
kind: List
items:
  - {{- include "namespace.yaml" . | nindent 4 }}
