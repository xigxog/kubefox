apiVersion: v1
kind: List
items:
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "service.yaml" . | nindent 4 }}
  - {{- include "deployment.yaml" . | nindent 4 }}
