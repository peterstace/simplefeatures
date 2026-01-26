# CLAUDE.md

@README.md explains what the packages in this directory are for.

@TRANSLITERATION_GUIDE.md describes how Java code is ported to Go code in those
packages.

The JTS repo will be at `../../locationtech/jts/` (relative to this file) and
have the tag that's being ported checked out. It is up to the human user to
ensure this is set up correctly.

## Workflow

- **One file per session**: Port one Java file per Claude Code session.  After
  completing a file, stop and wait for the human user to start a new session
  for the next file.

- When I request modifications to the ported code, also update
  `TRANSLITERATION_GUIDE.md` to reflect any new patterns that should be followed
  (to stop the same mistake being made again).

- **No third-party dependencies**: Do not use any third-party libraries,
  including for testing. Rely only on the Go standard library.
