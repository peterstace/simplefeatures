# JTS Port Review Approach

This document describes the systematic approach for reviewing ported files to
verify 1-1 correspondence with the original Java source.

## Goal

Verify that each ported Go file is a strict 1-1 **structural transliteration**
of its Java source, as defined in TRANSLITERATION_GUIDE.md.

This means:

- **1-1 file mapping**: Each Java file maps to exactly one Go file, and each Go
  file maps to exactly one Java file. Never combine multiple Java files into a
  single Go file. If during review you find a Go file that combines multiple
  Java files, it must be split before it can be marked as reviewed.

- **Line-by-line correspondence**: It should be possible to map between the
  Java and Go files line by line. Some lines may expand (one Java line → several
  Go lines) or contract (several Java lines → one Go line), but the structural
  mapping should be clear and traceable.

- **Preserved structure**: The order of methods, fields, and logic blocks should
  match the Java source. Don't reorder, reorganize, or "improve" the structure.

- **No extra code**: No additional methods, helper functions, or logic that
  doesn't exist in the Java source.

- **No missing code**: All Java methods, fields, and logic must be present in
  the Go version.

- **Equivalent logic**: The behavior must be identical, but this follows
  naturally from structural correspondence.

The goal is that a reviewer can read both files side-by-side and see the same
structure, making it easy to verify correctness and maintain the port as JTS
evolves.

## Progress Tracking

Progress is tracked in `MANIFEST.csv` using the `status` column:

| Status     | Meaning                                      |
| ---------- | -------------------------------------------- |
| `pending`  | Not yet ported                               |
| `ported`   | Ported but not yet reviewed                  |
| `reviewed` | Passed both LLM and human review             |

## Two-Step Review Process

Each file must pass **both** review steps before being marked as `reviewed`:

1. **LLM Review (Claude Code)**: The LLM performs initial structural comparison,
   flags issues, and fixes any problems found. The LLM does NOT update the
   status to `reviewed`.

2. **Human Review**: The human reviewer performs manual side-by-side comparison
   using vim. Only after the human confirms the file passes review should the
   status be updated to `reviewed`.

## Review Process

For each file:

1. **Locate the Java source** in `../../locationtech/jts/` (relative to the
   jtsport directory). The Java file path is in the first column of
   MANIFEST.csv.

2. **Open files side-by-side in vim:**
   ```
   vim -O <java-file> <go-file>
   ```
   The LLM reviewer should proactively provide this command for each file pair.

3. **Side-by-side structural comparison:**
   - Read through the Java and Go files together
   - Verify line-by-line correspondence is maintained
   - Check that methods appear in the same order
   - Verify fields are declared in the same order
   - Confirm control flow structures match (loops, conditionals, etc.)

3. **Check for completeness:**
   - All Java methods have corresponding Go methods (including unused/dead code)
   - All Java fields have corresponding Go fields
   - No extra methods or fields in Go that don't exist in Java
   - Static methods → package-level functions
   - Instance methods → receiver methods
   - Debug/print methods should be included even if they only produce output

4. **Check naming:**
   - Follows TRANSLITERATION_GUIDE.md conventions
   - Package prefixes are correct
   - Method names use Go conventions (exported vs unexported)
   - Static fields use full prefix: `javaPackage_className_memberName`

5. **Check test files:**
   - Uses `junit.AssertTrue`, `junit.AssertEquals`, etc. (not manual `if` + `t.Error`)
   - Imports `github.com/peterstace/simplefeatures/internal/jtsport/junit`

6. **Check polymorphism patterns:**
   - Child-chain dispatch pattern correctly implemented
   - `_BODY` methods where needed
   - `java.GetSelf()`, `java.InstanceOf[]`, `java.Cast[]` used appropriately

7. **Flag issues:**
   - Structural divergences (reordered methods, reorganized code)
   - Missing methods or fields
   - Extra methods or fields not in Java
   - Logic differences
   - Naming inconsistencies

8. **Fix issues** found during LLM review before human review begins.

9. **Human review**: After LLM review is complete, the human performs manual
   side-by-side comparison. Only after human approval should MANIFEST.csv be
   updated to `reviewed`.

## Review Session Format

Each LLM review session should:

1. State which phase/package is being reviewed
2. List the files to review in that batch
3. For each file, provide:
   - vim command for side-by-side viewing
   - Correspondence check (methods match)
   - Any issues found
   - Resolution (if issues were fixed)
4. Summarize files ready for human review (do NOT update MANIFEST.csv)

## Resuming Reviews

To continue a review in a new Claude Code session:

1. Check MANIFEST.csv for files with status `ported` (not yet reviewed)
2. Identify which phase those files belong to
3. Continue from the earliest incomplete phase
4. Follow the review process above

## Files Excluded from Review

The following are not subject to 1-1 review:

- `stubs.go` - Temporary stubs, will be replaced when dependencies are ported
- Test helper files that don't correspond to Java files
- Files in `xmltest/` - Test harness specific to Go

## Notes

- Test files (`*_test.go`) should be reviewed alongside their corresponding
  implementation files.
- The JTS repo should have tag v1.20.0 checked out for comparison.
