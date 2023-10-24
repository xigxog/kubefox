apiVersion: v1
kind: List
items:
  - {{- include "configmap-env.yaml" . | nindent 4 }}
  - {{- include "configmap-root-ca.yaml" . | nindent 4 }}
