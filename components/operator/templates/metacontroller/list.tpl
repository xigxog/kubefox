apiVersion: v1
kind: List
items:
  - {{- include "clusterrole.yaml" . | nindent 4 }}
  - {{- include "clusterrolebinding.yaml" . | nindent 4 }}
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "statefulset.yaml" . | nindent 4 }}
  - {{- include "controller.yaml" . | nindent 4 }}