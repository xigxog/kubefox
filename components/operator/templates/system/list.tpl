apiVersion: v1
kind: List
items:
  - {{- include "namespace.yaml" . | nindent 4 }}
  - {{- include "root-ca-secret.yaml" . | nindent 4 }}
  - {{- include "imagepullsecret-secret.yaml" . | nindent 4 }}
  - {{- include "broker-service.yaml" . | nindent 4 }}
  - {{- include "componentset.yaml" . | nindent 4 }}
