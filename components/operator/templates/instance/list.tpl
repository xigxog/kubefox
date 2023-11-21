apiVersion: v1
kind: List
items:
  - {{- include "validating-webhook.yaml" . | nindent 4 }}
