apiVersion: v1
kind: List
items:
  - {{- include "serviceaccount.yaml" . | nindent 4 }}
  - {{- include "clusterrole.yaml" . | nindent 4 }}
  - {{- include "clusterrolebinding.yaml" . | nindent 4 }}
  - {{- include "clusterrolebinding-auth.yaml" . | nindent 4 }}
  - {{- include "role.yaml" . | nindent 4 }}
  - {{- include "rolebinding.yaml" . | nindent 4 }}
  - {{- include "daemonset.yaml" . | nindent 4 }}
