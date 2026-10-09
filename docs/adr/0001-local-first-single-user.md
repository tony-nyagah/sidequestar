# Local-first, single-user app (no accounts, no server)

Sidequestar stores everything — quests, XP, photos, profile — in a local SQLite file on the user's machine, with no authentication and no hosted backend. We chose this over a multi-user hosted service because it keeps personal data off a server the user doesn't control, removes the entire auth/account surface, and lets one person run `go run` and be done. The trade-off is that there is no social feed or cross-device sync; for a tool whose point is to get you off screens, that's acceptable and likely permanent.
