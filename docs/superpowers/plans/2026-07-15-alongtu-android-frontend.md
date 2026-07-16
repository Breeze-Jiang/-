# Alongtu Android Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Execute tasks in dependency order and keep the checkbox state current.

**Goal:** Build the approved Alongtu Android front end as a stable, navigable React Native experience.

**Architecture:** The app uses a guest-first root stack with a five-tab main shell. Feature screens consume a typed local presentation model, while reusable UI components provide the approved token system, state presentation, and accessible touch targets. API adapters remain isolated from the navigation and mock presentation layers.

**Tech Stack:** React Native 0.76, TypeScript, React Navigation 6, Zustand, react-native-amap3d.

## Global Constraints

- Android is the launch platform; light mode is the approved first theme.
- Root navigation is Map, Discover, Messages, Favorites, Profile.
- Guests may browse, search, view details, and start navigation.
- Writes and social operations require login and must preserve the pending action.
- All primary touch targets are at least 48dp and use semantic labels.
- Use only data-backed wording for operating hours, existence, confirmation, and review state.

---

### Task 1: Replace legacy app shell

**Files:** `place_app/src/navigation/RootNavigator.tsx`, `place_app/src/shared/utils/theme.ts`, `place_app/src/app/*`

- [ ] Build the guest-first root stack and five-tab shell.
- [ ] Create semantic theme tokens and reusable headers, status pills, buttons, states and cards.
- [ ] Verify that an unauthenticated launch shows the main app rather than an auth screen.

### Task 2: Implement travel convenience path

**Files:** `place_app/src/app/screens/MapHomeScreen.tsx`, `PlaceResultsScreen.tsx`, `PlaceDetailScreen.tsx`, `NavigationScreen.tsx`

- [ ] Provide always-visible urgent service entries and an expanded Alongtu service panel.
- [ ] Render distance-first results with information provenance and a direct route action.
- [ ] Provide travel mode selection, driving mode, and recoverable permission/network states.

### Task 3: Implement discovery, assets and travel planning

**Files:** `place_app/src/app/screens/DiscoverScreen.tsx`, `AssetScreens.tsx`, `CreationScreens.tsx`

- [ ] Build recommendation/public community navigation and its three public channels.
- [ ] Build collection folders, private markers, plan cards and creation routes.
- [ ] Make public marker/plan intent visibly distinct from private drafts and show review states.

### Task 4: Implement social, identity and governance routes

**Files:** `place_app/src/app/screens/SocialScreens.tsx`, `AccountScreens.tsx`, `src/store/*`

- [ ] Build message, friend, chat, group and voluntary location-sharing surfaces.
- [ ] Add login gate with pending-action continuation model.
- [ ] Build notification, review and report states with explicit explanations.

### Task 5: Validate and audit

**Files:** `place_app/package.json`, `docs/ALONGTU_MASTER_PLAN.md`, `docs/FRONTEND_COVERAGE_AUDIT.md`

- [ ] Run TypeScript, lint, unit tests and Android build once dependencies are installed.
- [ ] Check route reachability and empty/error/guest states for every approved module.
- [ ] Record implemented UI coverage without marking backend capabilities as complete.
