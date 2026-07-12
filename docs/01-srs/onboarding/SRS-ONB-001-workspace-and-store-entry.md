# SRS-ONB-001 Workspace And Store Entry

Parent: [Podzone SRS](../podzone-srs.md) · [Traceability Matrix](../traceability-matrix.md)

The system shall require users to choose a workspace and then a ready store
before entering store-scoped Backoffice operations. A workspace with no store
yet shall present store-request submission as the entry action, not an empty
or generic chooser screen; a workspace with one or more stores shall present
selection among them, distinguishing pending, failed, and ready state per
store. See `../../06-recovery/backbone-flow-refactor.md` "Required Screens"
(Workspace/store chooser) for the exact state list.

Linked docs:

- `../../00-project-vision/09-backoffice-multitenancy.md`
- `../../00-project-vision/10-store-onboarding-pipeline.md`
