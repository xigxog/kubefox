apiVersion: v1
kind: List
items:
  - {{- include "configmap-env.yaml" . | nindent 4 }}
  - {{- include "configmap-pki-init.yaml" . | nindent 4 }}
