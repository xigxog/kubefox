apiVersion: v1
kind: List
items:
  - {{- include "mutating-webhook.yaml" . | nindent 4 }}
  - {{- include "validating-webhook.yaml" . | nindent 4 }}
