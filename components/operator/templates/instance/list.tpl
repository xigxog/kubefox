apiVersion: v1
kind: List
items:
  - {{- include "configmap-env.yaml" . | nindent 4 }}
  - {{- include "mutating-webhook.yaml" . | nindent 4 }}
  - {{- include "validating-webhook.yaml" . | nindent 4 }}
