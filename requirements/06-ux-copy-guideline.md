# UX Copy Guideline

## Purpose

This document defines how Podzone should speak to users if it is positioned as a `POD-first operations platform`.

The current UI still uses many technical or infrastructure-heavy terms such as:

- tenant
- membership
- platform role
- active tenant
- IAM

These terms are acceptable for engineering and internal docs, but they should not dominate the merchant-facing experience.

## Core writing principles

## 1. Use business language first

Prefer words that merchants and operators expect.

Good:

- store
- team
- print partner
- products
- orders
- fulfillment
- margin

Avoid:

- tenant
- actor
- principal
- membership binding

## 2. Speak in actions, not system mechanisms

Good:

- Open store
- Invite teammate
- Connect print partner
- Set up products
- Review orders

Avoid:

- Switch active tenant
- Add membership
- Bind role
- Resolve identity

## 3. Keep infrastructure invisible unless the user truly needs it

Users need to think about:

- which store they are in
- who has access
- which partner is connected
- which orders need attention

## Product vocabulary

### Preferred replacements

- `Tenant` -> `Store` or `Workspace`
- `Tenant ID` -> `Store ID`
- `Membership` -> `Store access` or `Team access`
- `Tenant member` -> `Store teammate`
- `Tenant owner` -> `Store owner`
- `Tenant admin` -> `Store admin`
- `Tenant editor` -> `Store operator`
- `Platform role` -> `Platform admin role`
- `Tenant invite` -> `Store invite`
- `Accept tenant invite` -> `Join store team`
- `Switch active tenant` -> `Open store`

### POD-specific vocabulary

- `Print partner`
- `Production partner`
- `Product setup`
- `Publish to store`
- `Fulfillment status`
- `Shipment tracking`
- `Margin`
- `Payout`

### Extensible vocabulary

When the platform later broadens beyond POD:

- `Partner` can become the umbrella term
- `Print partner` remains the default user-facing wording in POD flows
- `Supplier` should only appear in dropship-specific surfaces

## UI area guidance

## 1. Auth

Preferred examples:

- `Sign in to manage your stores`
- `Sign in with Google`
- `Join your store team`

## 2. Admin home

Preferred labels:

- `My stores`
- `Create store`
- `Open store`
- `Current store`
- `Team access`

## 3. Settings

Preferred sections:

- `Team access`
- `Store invites`
- `Platform administration`
- `Sessions`
- `Activity log`

## 4. Partner-facing future screens

Preferred sections:

- `Print partners`
- `Production settings`
- `Catalog setup`
- `Fulfillment performance`
- `Partner issues`

## 5. Orders

Preferred labels:

- `Orders`
- `Needs attention`
- `Waiting for fulfillment`
- `Fulfillment delayed`
- `Partially fulfilled`

## Example rewrites

- `Provision a new workspace and attach yourself as tenant owner.` -> `Create a new store and start managing products, team access, and orders.`
- `Connect supplier` -> `Connect print partner`
- `Supplier issue` -> `Partner issue`

## BA recommendation

The immediate goal is:

`Make Podzone feel like a seller operations product for POD businesses, not like a tenant administration console.`
