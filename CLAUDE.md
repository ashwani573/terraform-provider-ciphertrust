# terraform-provider-ciphertrust

Terraform provider for **Thales CipherTrust Manager** (CM) and **CDSPaaS**. Built on `terraform-plugin-framework` (NOT SDKv2).

## Build / test
- `make build` / `make test` / `make testacc` (live CM, `TF_ACC=1` + `CIPHERTRUST_*` env vars) / `make lint` / `make fmt` / `make generate` (regen `docs/`)

## Top-level layout
- [main.go](main.go) → [internal/provider/provider.go](internal/provider/provider.go) (registration; **all new resources go here**)
- Subsystems: `cm/`, `connections/`, `cckm/{aws,oci,acls,mutex,utils}/`, `cte/`, `common/`, `models/` — full map in [.claude/indexes/subsystems.md](.claude/indexes/subsystems.md)
- `docs/` is **generated** — edit `templates/` + `examples/`, then `make generate`

## Pre-built indexes — read these instead of grepping
- [.claude/indexes/resources.md](.claude/indexes/resources.md) — every resource → file:line
- [.claude/indexes/data-sources.md](.claude/indexes/data-sources.md) — every data source → file:line
- [.claude/indexes/subsystems.md](.claude/indexes/subsystems.md) — what lives where
- [.claude/indexes/conventions.md](.claude/indexes/conventions.md) — resource skeleton, `*common.Client` CRUD helpers, URL constants, logging, 404 behavior
- [.claude/indexes/new-resource-recipe.md](.claude/indexes/new-resource-recipe.md) — step-by-step for adding a resource
- Subagents in [.claude/agents/](.claude/agents/): `ciphertrust-locator` (fast find), `ciphertrust-resource-author` (add new resource end-to-end)

## Swagger / API reference
The CipherTrust Manager API spec is at [definition-beta.json](definition-beta.json) — **~14.8 MB / ~3.7M tokens. DO NOT load directly.** Pre-split + deduplicated files live under [.claude/swagger/](.claude/swagger/):
- [.claude/swagger/index.md](.claude/swagger/index.md) — area map (3.7 KB)
- [.claude/swagger/operations.md](.claude/swagger/operations.md) — grep-searchable line-per-operation TOC (~260 KB; grep it, don't read it)
- [.claude/swagger/areas/<area>.json](.claude/swagger/areas/) — per-area swagger (16–430 KB each) — references shared schemas via `"$ref": "../definitions.json#/D####"`
- [.claude/swagger/definitions.json](.claude/swagger/definitions.json) — 494 shared schemas (~635 KB). **Don't load whole — grep for the specific `D####` name and read just those lines.**
- Regenerate: `python .claude/swagger/scripts/regenerate.py`

## Must-knows (don't violate without thinking)
- TF type name: `req.ProviderTypeName + "_<name>"` → `ciphertrust_<name>`
- Use `*common.Client` helpers (`GetById`/`PostData`/`UpdateData`/`DeleteByURL`/…), NOT raw `http.NewRequest`. URL constants in [common/urls.go](internal/provider/common/urls.go).
- **On Read 404: keep resource in state** unless mid-Delete (commit `43f3b14`, TFIN-185).
- Active branch is **`1.0.1`**, not `main`. PRs target `1.0.1`.

Update [.claude/indexes/resources.md](.claude/indexes/resources.md) / [data-sources.md](.claude/indexes/data-sources.md) when adding/renaming a constructor.
