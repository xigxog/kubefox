apiVersion: v1
kind: List
items:
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "clusterrole.yaml" . | nindent 4 }}
  - {{- include "clusterrolebinding.yaml" . | nindent 4 }}
  - {{- include "service.yaml" . | nindent 4 }}
  - {{- include "broker-service.yaml" . | nindent 4 }}
  - {{- include "deployment.yaml" . | nindent 4 }}
