# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Terraform provider `lablabs/lara-utils` — utility **provider functions** only (no resources/data sources), built on **terraform-plugin-framework v1** (protocol v6), NOT the legacy SDKv2. Two functions: `deep_merge` and `yaml_deep_merge`, called as `provider::lara-utils::deep_merge(objects, options...)`.

## Commands

Use **mise** for everything — there is no Makefile.

- `mise run build` — `go build -v ./...`
- `mise run test` — unit tests (`go test -v -cover -timeout=120s -parallel=10 ./...`)
- `mise run testacc` — acceptance tests (sets `TF_ACC=1`; runs against `./internal/provider/`)
- `mise run lint` — `golangci-lint run`
- `mise run fmt` — `gofmt -s -w -e .`
- `mise run generate` — regenerate docs & license headers (`cd tools && go generate ./...`)
- `mise run` (default) — `fmt, lint, install, generate`; run before submitting a PR.

Run a single test: `go test ./internal/provider/ -run TestName -v`.

## Gotchas

- **Docs & license headers are generated and CI-enforced.** After changing functions, examples, or comments, run `mise run generate` and commit the result — CI runs `git diff --exit-code` and fails otherwise. Generation runs `copywrite headers`, `terraform fmt`, and `tfplugindocs`.
- Function descriptions are `//go:embed`-ed from `internal/provider/*_function.md` — edit those `.md` files to change function docs, not the generated `docs/`.
- `tools/` is a **separate Go module** (own `go.mod`) holding codegen tools — don't add codegen deps to the root `go.mod`.
- golangci-lint **v2** config, `default: none`. `godot` requires comments to end with a period; `forcetypeassert` is on, so deliberate assertions use `//nolint:forcetypeassert`.
- Core merge logic lives in `internal/deepmerge/` (wraps `dario.cat/mergo`); `attr.Value` ⇄ Go `any` bridging lives in `internal/helpers/`. Shared test-step builders are in `internal/provider/testdata/`.

## Conventions

- Commits follow **Conventional Commits** (`feat:`, `fix:`, `chore:`, `docs:`, etc.).
- License MPL-2.0; copyright holder "Labyrinth Labs s.r.o." (some scaffold-derived files still carry HashiCorp headers).
