apiVersion: v1
kind: List
items:
  - {{- include "configmap-root-ca.yaml" . | nindent 4 }}
