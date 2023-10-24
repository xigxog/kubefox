apiVersion: v1
kind: List
items:
  - {{- include "configmap-env.yaml" . | nindent 4 }}
