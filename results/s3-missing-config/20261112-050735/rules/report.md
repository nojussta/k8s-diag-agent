Incident triage: LabPodNotReady
Probable cause: ConfigurationError in demo/deployment/frontend (confidence high)
container frontend cannot be created (CreateContainerConfigError: configmap "frontend-config" not found). It started with the newest revision: revision 56 (change-cause: release 2.1: frontend configuration update) changed frontend env: [{"name":"PODINFO_UI_MESSAGE","value":"frontend"}] -> [{"name":"PODINFO_UI_MESSAGE","value":"frontend"},{"name":"PODINFO_UI_COLOR","valueFrom":{"configMapKeyRef":{"key":"ui.color","name":"frontend-config"}}}].

Evidence:
- [A2 pod status] phase=Pending; frontend restarts=0 state=CreateContainerConfigError
- [A3 events] 2026-11-12T03:09:54Z Warning Failed x12: Error: configmap "frontend-config" not found
- [A5 rollout history] revision 56 (change-cause: release 2.1: frontend configuration update) changed frontend env: [{"name":"PODINFO_UI_MESSAGE","value":"frontend"}] -> [{"name":"PODINFO_UI_MESSAGE","value":"frontend"},{"name":"PODINFO_UI_COLOR","valueFrom":{"configMapKeyRef":{"key":"ui.color","name":"frontend-config"}}}].

Suggested next steps (not executed; a person decides):
- Roll back to the previous revision: kubectl -n demo rollout undo deployment/frontend

Method rules, model none, 19 read-only calls, 0.1 s.
