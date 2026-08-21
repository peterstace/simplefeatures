# CLAUDE.md

Update CHANGELOG.md when a change alters what callers of this module can
observe. That covers API that is added, removed, or renamed, changed
behaviour or results, bug fixes, and performance work. Notable internal work
can also be logged, saying that it is not externally detectable.

Skip the entry when nothing a caller could depend on has changed, such as the
wording of an error message, comments, or test-only changes. Where a change
refines something already listed under Unreleased, extend that entry rather
than adding a second entry.
