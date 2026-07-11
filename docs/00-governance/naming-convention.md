# Naming Convention

## Files And Folders

Use lowercase kebab-case for documentation folders and files.

Good:

```text
api-protection/
functional-requirements.md
business-rules.md
```

Avoid:

```text
ApiProtection/
FunctionalRequirements.md
03-Usecases/
```

Existing internal terms such as TKTT/TKCT may remain when they are already part
of the project vocabulary.

## Document IDs

Use stable IDs:

| Prefix | Meaning |
|---|---|
| `BRD-<DOMAIN>-NNN` | Business requirement |
| `UC-<DOMAIN>-NNN` | Use case |
| `FR-<DOMAIN>-NNN` | Functional requirement |
| `NFR-<DOMAIN>-NNN` | Non-functional requirement |
| `BR-<DOMAIN>-NNN` | Business rule |
| `AC-<DOMAIN>-NNN` | Acceptance criteria |
| `PZEP-NNNN` | Podzone Enhancement Proposal |
| `ADR-NNNN` | Architecture Decision Record |
| `TASK-NNNN` | Implementation task |
| `SCREEN-<DOMAIN>-NNN` | UI screen spec |

Examples:

```text
FR-ONBOARDING-001
UC-IAM-001
AC-BACKOFFICE-001
PZEP-0001
ADR-0002
TASK-0008
```
