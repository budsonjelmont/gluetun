# MozillaVPN provider implementation plan

## Purpose

This document captures the implementation plan for adding MozillaVPN support to Gluetun in a way that can be committed to the repository and resumed from another development environment.

The intended outcome is one of:

1. A built-in `mozillavpn` provider implemented in Gluetun.
2. A documented fallback to custom-provider guidance if MozillaVPN cannot satisfy Gluetun's built-in provider constraints.

## Scope

### In scope

- Researching MozillaVPN protocol and server-data feasibility.
- Deciding whether a built-in provider is valid under the Gluetun wiki rules.
- Implementing provider registration, updater logic, runtime selection, validation rules, documentation entries, and tests if feasible.
- Producing a clear checklist for generating provider server data.

### Out of scope for the first pass

- Opening upstream pull requests.
- Maintaining a fork of `gluetun-servers` long term.
- Adding support for protocols that MozillaVPN does not reliably expose to Gluetun.
- Adding speculative features with no verified Mozilla source.

## Current decisions

### User decisions already made

- Provider display name preference: `MozillaVPN`
- Intended protocol direction: WireGuard only
- Requested implementation scope: functional implementation in this repository
- Fallback policy if built-in provider is not valid: switch to custom-provider guidance and tooling only

### Internal naming recommendation

Use:

- user-facing name: `MozillaVPN`
- provider constant value: `mozillavpn`
- provider package directory: `internal/provider/mozillavpn`
- generated servers file name: `mozillavpn.json`

This follows the Gluetun wiki guidance that the provider string should be lowercase without spaces or underscores.

## Governing constraints

### Constraint from the Gluetun provider wiki

A built-in WireGuard provider is only appropriate if the provider data model is compatible with Gluetun's static provider model.

Important rule from the wiki:

- If WireGuard `PrivateKey` and `Address`, plus any `PreSharedKey`, are different for each server, the user should use the custom provider and Gluetun should not implement built-in WireGuard support for that provider.
- If the user interface material is stable across servers and the server public keys are either constant or derivable, then a built-in provider can be considered.

### Practical implication

The first implementation gate is not code. It is proving whether MozillaVPN can be represented as a built-in Gluetun provider without requiring per-server, user-specific configuration generation.

If that proof fails, the correct implementation outcome is not partial provider code. It is a documented custom-provider workflow.

## Current evidence gathered

### Confirmed

- MozillaVPN is WireGuard-focused.
- MozillaVPN requires Mozilla account sign-in in the official client.
- The Mozilla VPN client repository contains a CI script using Mullvad's public WireGuard relay API:
  - `https://api.mullvad.net/public/relays/wireguard/v2`
- That script appears to use the Mullvad relay data at least for country and city naming updates.

### Not yet confirmed

- Whether MozillaVPN runtime connection data can be obtained from a public API suitable for Gluetun.
- Whether endpoint IPs, endpoint ports, and WireGuard public keys used by MozillaVPN are identical to or derivable from the Mullvad public relay API.
- Whether MozillaVPN requires authenticated API access to retrieve user-specific WireGuard interface details.
- Whether the interface address, private key, or preshared key are per-user only, or per-user and per-server.
- Whether any Mozilla-specific restrictions exist beyond Mullvad relay metadata.

### Working hypothesis

The most promising path is that MozillaVPN may be mappable as a WireGuard-only provider by reusing a public relay source, likely related to Mullvad relay metadata, plus standard Gluetun WireGuard selection behavior.

That hypothesis is still unproven and must be validated before significant provider code is written.

## Feasibility gate

Implementation should only proceed to full provider coding if all of the following are answered satisfactorily.

### Gate A: server inventory source

Need a reliable source for at least:

- server hostname or name
- endpoint IP or resolvable hostname
- country, region, city when available
- WireGuard public key
- enough stable metadata to generate `models.Server` entries

Acceptable sources:

- public JSON API
- provider repository source with stable consumable metadata
- static data file shipped by Mozilla if license and update cadence are acceptable

Unacceptable source:

- manual copy-paste list with no repeatable update path

### Gate B: WireGuard interface compatibility

Need to determine whether built-in provider semantics are valid.

Built-in provider is viable only if:

- the user is not required to supply a different interface private key or address for each selected server
- or the Gluetun model can support the exact runtime data without violating the wiki rules

Built-in provider is not viable if:

- each server selection requires a server-specific user-issued WireGuard interface config
- a per-server preshared key is mandatory and cannot be derived from stable metadata
- the official service model is essentially "download a bespoke WireGuard config for each chosen server"

### Gate C: authentication model

Need to determine whether Gluetun would require new credentials or tokens.

Possible outcomes:

1. No extra auth needed for server inventory and server public keys.
2. Auth needed only for account login, but not for server inventory.
3. Auth needed for runtime WireGuard configuration material.

If outcome 3 is true, built-in provider feasibility becomes much lower and may be invalid.

### Gate D: settings fit

Need to confirm that MozillaVPN can be expressed using the current settings surface or with a narrowly-scoped addition.

Prefer:

- no new environment variables
- reuse of existing WireGuard settings where possible

Avoid unless clearly justified:

- adding multiple Mozilla-specific settings
- adding token exchange flows inside Gluetun without a stable documented API

## Implementation phases

## Phase 1: external research and proof

This phase blocks all meaningful provider coding.

### Goals

- identify authoritative MozillaVPN connection-data sources
- verify whether a Gluetun built-in provider is valid
- record exact evidence and URLs

### Tasks

1. Inspect the Mozilla VPN client repository for:
   - relay list consumption
   - WireGuard endpoint configuration generation
   - account login or token usage for connection provisioning
   - any references to endpoint IP, endpoint port, server public key, or interface address generation
2. Check whether MozillaVPN is simply using Mullvad relay infrastructure and whether the public Mullvad relay API is sufficient for runtime server metadata.
3. Determine whether user-specific interface data is fetched at login or baked into a generic WireGuard profile.
4. Record one of these conclusions:
   - built-in provider is feasible
   - built-in provider is not feasible, use custom-provider guidance
   - more evidence required before coding

### Deliverable

A short evidence-backed decision note added to this file or a follow-up section.

## Phase 2: provider registration scaffold

Only do this if Phase 1 concludes built-in provider is feasible.

### Files to modify

- `internal/constants/providers/providers.go`
- `internal/provider/providers.go`

### Tasks

1. Add provider constant in alphabetical order:

```go
MozillaVPN = "mozillavpn"
```

2. Add the provider to `All()` in alphabetical order.
3. Import the provider package in `internal/provider/providers.go`.
4. Add the provider to the `providerNameToProvider` map in alphabetical order.
5. Ensure the registration count sanity check still matches `providers.AllWithCustom()`.

### Notes

The provider constructor signature should match actual feature needs. For a WireGuard-only implementation it will likely resemble Mullvad more than the example provider.

## Phase 3: provider package creation

Only do this if Phase 1 concludes built-in provider is feasible.

### Create directory

- `internal/provider/mozillavpn`

### Starting point

Use these references:

- template structure: `internal/provider/example`
- WireGuard-only reference: `internal/provider/mullvad`

### Expected files

At minimum:

- `internal/provider/mozillavpn/provider.go`
- `internal/provider/mozillavpn/connection.go`
- `internal/provider/mozillavpn/openvpnconf.go`
- `internal/provider/mozillavpn/updater/updater.go`
- `internal/provider/mozillavpn/updater/servers.go`
- additional updater helpers as needed such as `api.go`

### Package behavior

- `Name()` must return `providers.MozillaVPN`
- `GetConnection()` should use WireGuard defaults appropriate for MozillaVPN
- `OpenVPNConfig()` should explicitly reject or panic if OpenVPN is unsupported, following the pattern used by other unsupported-provider combinations
- `Fetcher` should be implemented through the updater package

## Phase 4: updater implementation

This is the core of the provider.

### Goal

Produce valid `[]models.Server` entries for MozillaVPN.

### Required per server

For WireGuard server entries, ensure:

- `VPN` is set to `wireguard`
- `Hostname` or `ServerName` is set
- `IPs` contains one or more endpoint IPs
- `WgPubKey` is populated

### Desired optional fields

- `Country`
- `Region`
- `City`
- `ISP` if meaningful
- `Owned` only if semantically correct

### Updater tasks

1. Define source structs matching the chosen external source.
2. Fetch the source data using a context-aware HTTP client.
3. Validate minimum server count.
4. Normalize raw source data into Gluetun `models.Server` values.
5. If the source gives hostnames but not IPs, resolve them using the shared parallel resolver if needed.
6. Sort servers deterministically using `models.SortableServers`.
7. Return enough structured data for `go run ./cmd/gluetun/main.go update -providers mozillavpn` to generate server data.

### Key design choice

Prefer consuming endpoint IPs directly from source data if available. Avoid DNS resolution if the upstream data already contains canonical WireGuard endpoint IPs.

### Likely pitfalls

- mapping a location source that is not actually the runtime endpoint source
- missing server public keys
- using Mullvad metadata that does not correspond to MozillaVPN-allowed relay set
- building an updater that works once but has no stable re-run path

## Phase 5: runtime connection behavior

### Reference

Model this after WireGuard-only providers, especially Mullvad, not after OpenVPN-oriented providers.

### Connection defaults

Update `internal/provider/mozillavpn/connection.go` to use:

- OpenVPN TCP default port: `0` if unsupported
- OpenVPN UDP default port: `0` if unsupported
- WireGuard default port: actual MozillaVPN default if one is fixed, otherwise a reasonable placeholder consistent with source metadata strategy

Do not guess a fixed port if the provider's server data already carries endpoint IP plus server-specific port logic elsewhere.

### OpenVPN behavior

If MozillaVPN is WireGuard only, `openvpnconf.go` should not attempt to build an OpenVPN config.

Recommended options:

1. explicit panic with a clear message indicating MozillaVPN OpenVPN is unsupported
2. mirror the pattern used by unsupported protocol/provider combinations already in the repo

The chosen behavior should match surrounding project conventions.

## Phase 6: settings validation changes

### Files to review

- `internal/configuration/settings/openvpnselection.go`
- `internal/configuration/settings/wireguardselection.go`
- possibly `internal/configuration/settings/openvpn.go`

### OpenVPN selection changes

If MozillaVPN is WireGuard only:

- add `providers.MozillaVPN` to the list of providers where OpenVPN TCP is unsupported if applicable
- ensure OpenVPN UDP is also rejected where necessary through the appropriate validation path
- ensure unsupported protocol selection fails early and clearly

### WireGuard selection changes

If MozillaVPN is supported as a built-in WireGuard provider:

- add `providers.MozillaVPN` to the supported-provider branch in `wireguardselection.go`
- determine whether `EndpointPort` should be:
  - required
  - optional and unrestricted
  - optional but restricted to a fixed set
  - forbidden because the provider data already fixes it
- determine whether `PublicKey` is baked in or not

### Important note

`wireguardselection.go` behavior differs by provider regarding endpoint port and public key handling. MozillaVPN must be added based on actual data shape, not by analogy alone.

## Phase 7: markdown and provider listings

### Files to modify

- `internal/models/markdown.go`
- `.github/ISSUE_TEMPLATE/bug.yml`
- `.github/labels.yml`
- `README.md`

### Tasks

1. Add `providers.MozillaVPN` to markdown header selection in `internal/models/markdown.go`.
2. Choose a header set consistent with actual server fields. Likely candidates:
   - country
   - city or region
   - hostname
   - vpn
3. Add `MozillaVPN` to the provider dropdown in `.github/ISSUE_TEMPLATE/bug.yml` in alphabetical order.
4. Add a provider label entry to `.github/labels.yml` in alphabetical order.
5. Add `MozillaVPN` to the supported providers line in `README.md`.

### Documentation nuance

If the provider ends up being WireGuard only, the README entry should say so, matching the style used for Mullvad.

Example style:

- `**MozillaVPN** (Wireguard only)`

Maintain the exact capitalization style already used in the README.

## Phase 8: tests

### Goal

Add enough tests to protect the provider-specific behavior.

### Candidate test areas

1. provider registration sanity if touched indirectly
2. updater parsing and normalization
3. `GetConnection()` defaults
4. unsupported OpenVPN behavior if implemented explicitly
5. settings validation around WireGuard and OpenVPN protocol restrictions

### Test patterns to follow

- map-based table tests
- `t.Parallel()` at top-level and subtest level
- use concrete expectations, no loose matching

### Likely locations

- `internal/provider/mozillavpn/*_test.go`
- existing settings test files if protocol validation coverage is extended

## Phase 9: validation commands

Run these before considering the implementation complete.

### Required project validation

```sh
go build ./...
golangci-lint run
go test ./...
```

### Provider-specific validation

```sh
go run ./cmd/gluetun/main.go update -providers mozillavpn
```

### Optional markdown validation

If available in the environment:

```sh
markdownlint-cli2 "**/*.md"
```

### Validation outcomes to inspect manually

- provider count sanity check does not panic
- generated server data contains only valid WireGuard entries
- generated server data has stable location fields
- there are no empty required fields such as missing endpoint IPs or missing public keys
- protocol selection errors are clear and correct

## Detailed file checklist

## Existing files expected to change

### Core provider registration

- `internal/constants/providers/providers.go`
- `internal/provider/providers.go`

### Settings validation

- `internal/configuration/settings/openvpnselection.go`
- `internal/configuration/settings/wireguardselection.go`
- possibly `internal/configuration/settings/openvpn.go`

### Markdown and docs

- `internal/models/markdown.go`
- `.github/ISSUE_TEMPLATE/bug.yml`
- `.github/labels.yml`
- `README.md`

### New provider files likely to be added

- `internal/provider/mozillavpn/provider.go`
- `internal/provider/mozillavpn/connection.go`
- `internal/provider/mozillavpn/openvpnconf.go`
- `internal/provider/mozillavpn/updater/updater.go`
- `internal/provider/mozillavpn/updater/servers.go`
- `internal/provider/mozillavpn/updater/api.go`
- one or more test files under `internal/provider/mozillavpn`

## Acceptance criteria

The work is complete only if all relevant statements below are true.

### Feasibility

- there is documented evidence that MozillaVPN is representable as a Gluetun built-in provider
- or the document clearly states why it is not, and the effort pivots to a custom-provider path

### Code integration

- `VPN_SERVICE_PROVIDER=mozillavpn` is recognized
- provider registration succeeds without mismatch panics
- WireGuard provider selection validates correctly
- unsupported protocol selections fail clearly

### Server updater

- updater fetches repeatable provider data from a documented source
- server entries contain required fields
- provider update command succeeds

### Quality

- build passes
- lint passes
- tests pass
- documentation/provider lists are updated consistently

## Risks and decision points

### Risk 1: MozillaVPN is not a valid built-in provider

This is the highest risk.

If the official client requires per-server user-issued WireGuard configs, stop built-in implementation and switch to custom-provider guidance.

### Risk 2: public server metadata is incomplete

The Mullvad relay API may expose location names and WireGuard relay data, but MozillaVPN may not allow the full Mullvad relay set or may use additional policy not reflected there.

### Risk 3: account-bound provisioning

Mozilla account authentication may be required to derive device- or account-bound interface details. If so, Gluetun may need a design the current provider model does not support.

### Risk 4: overfitting to one current upstream source

If implementation depends on a fragile Mozilla internal endpoint or an unstable repo artifact, the updater may be too brittle for long-term maintenance.

## Fallback plan if built-in provider is invalid

If Phase 1 concludes a built-in provider is not correct, the replacement deliverable should be:

1. keep this document and update it with the feasibility conclusion
2. do not add partial `mozillavpn` provider code
3. write documentation explaining how MozillaVPN users can use Gluetun's custom WireGuard provider instead
4. if useful, add helper documentation or scripts outside the core provider model only if they fit repository standards

That outcome is still a successful result because it follows the project wiki and avoids shipping a misleading provider integration.

## Suggested next actions from a fresh environment

1. Re-check this file and confirm the feasibility gate is still unresolved.
2. Start with external evidence gathering, not code edits.
3. Capture exact Mozilla source locations and API shapes.
4. Only after that, choose one of:
   - proceed with built-in provider implementation
   - pivot to custom-provider guidance

## Research log template

Use this section to continue work in another environment.

### Evidence entries

- Date:
- Source:
- URL or file path:
- What it proves:
- Impact on feasibility:

### Decision entries

- Date:
- Decision:
- Reason:
- Next step:

## Status at time of writing

- wiki guidance reviewed
- repository integration points mapped
- high-level implementation path identified
- external feasibility still unresolved
- no provider code committed yet

## Short resume guide

If resuming later, begin here:

1. verify MozillaVPN connection model
2. decide built-in vs custom-provider path
3. if built-in is valid, implement registration and updater first
4. then wire validation, docs, tests, and provider update verification
