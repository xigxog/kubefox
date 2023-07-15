apiVersion: v1
kind: List
items:
  - {{- include "componentset.yaml" . | nindent 4 }}
  - {{- include "compositecontroller.yaml" . | nindent 4 }}
  - {{- include "controllerrevision.yaml" . | nindent 4 }}
  - {{- include "decoratorcontroller.yaml" . | nindent 4 }}
  - {{- include "ingressroute.yaml" . | nindent 4 }}
  - {{- include "middleware.yaml" . | nindent 4 }}
  - {{- include "platform.yaml" . | nindent 4 }}
  - {{- include "release.yaml" . | nindent 4 }}
