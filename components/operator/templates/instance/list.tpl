apiVersion: v1
kind: List
items:
  - {{- include "configmap-env.yaml" . | nindent 4 }}
  - {{- include "webhook.yaml" . | nindent 4 }}
