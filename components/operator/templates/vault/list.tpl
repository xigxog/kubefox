apiVersion: v1
kind: List
items:
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "role.yaml" . | nindent 4 }}
  - {{- include "rolebinding.yaml" . | nindent 4 }}
  - {{- include "clusterrolebinding.yaml" . | nindent 4 }}
  - {{- include "entrypoint-configmap.yaml" . | nindent 4 }}
  - {{- include "service.yaml" . | nindent 4 }}
  - {{- include "statefulset.yaml" . | nindent 4 }}
