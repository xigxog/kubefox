apiVersion: v1
kind: List
items:
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "configmap.yaml" . | nindent 4 }}
  - {{- include "service.yaml" . | nindent 4 }}
  - {{- include "statefulset.yaml" . | nindent 4 }}
