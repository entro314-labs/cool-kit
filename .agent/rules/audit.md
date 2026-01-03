---
trigger: model_decision
description: use it when auditing the codebase
---

# Code Audit Directives

## Core Principle

Before adding any code, verify no equivalent functionality exists in the codebase. Prefer wiring and composing existing implementations over introducing new abstractions. Duplication is a defect.

## Error Handling

- Flag silent failures—operations that can fail but don't surface errors
- Flag unchecked errors—return values ignored, promises without catch/rejection handling
- Flag error swallowing—catch blocks that log-and-continue or return generic fallbacks without propagation context
- Flag missing error wrapping—errors re-thrown without context or stack preservation
- Require explicit error paths for all fallible operations

## Type Safety

- Flag `any`, `unknown` casts, and implicit `any` in TypeScript
- Flag `type: ignore` comments in Python
- Flag untyped function signatures and return types
- Flag nil/null hazards—nullable access without guards
- Flag unsafe type assertions that bypass compiler checks
- Flag non-exhaustive switch/match statements—all variants must be handled or explicitly unreachable

## Language-Specific Hazards

**Rust:**
- Flag `.unwrap()`, `.expect()`, and `unsafe` blocks outside test code
- Require `?` propagation or explicit error handling

**Python:**
- Flag bare `except:` clauses—must specify exception types
- Flag broad `except Exception` without re-raise

**Go:**
- Flag `_ = err` patterns
- Flag unchecked error returns

## Implementation Integrity

- Flag stubs, mocks, TODO implementations, and placeholder logic outside test directories
- Flag dead code—unreachable branches, unused functions, commented-out blocks
- Flag partial implementations—functions that handle happy path only
- Flag broken invariants—state that can violate documented or implied contracts
- Flag orphaned functionality—code with no call sites or integration points

## Framework Boundaries (Next.js / React)

- Flag server/client boundary leaks—server-only imports in client components, client state in server components
- Flag misplaced `"use client"` directives—should be at the top of files that need client-side execution
- Flag hydration mismatches—initial render divergence between server and client

## Database & Schema Alignment

- Flag missing foreign key indexes—every FK column needs an index
- Flag constraint gaps—application logic enforcing rules that should be database constraints
- Flag type changes that don't align with schema migrations
- Flag orphaned migrations or schema drift

## Technical Debt Markers

- Flag `// TODO`, `// FIXME`, `// HACK` comments older than the current work scope
- Flag duplicated logic across files—extract or consolidate
- Flag inconsistent patterns—same problem solved differently in different places

## Action Protocol

- **Fix** issues where the correct resolution is unambiguous and low-risk
- **Flag** issues where multiple valid approaches exist or changes have broad impact
- **Document** the specific location, nature of the issue, and recommended resolution for all flags