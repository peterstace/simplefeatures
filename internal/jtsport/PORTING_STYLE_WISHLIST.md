# Porting Style Wishlist

This document describes some changes to the porting style that I'd like, but
aren't critical to the first cut of the port. They are either stylistic, or at
the very least hidden behind the package interface.

## Review code for 1-1 correspondence with Java

The intent was to be 1-1 from the start, but this may have drifted and should
be confirmed systematically and fixed where appropriate.

## Filename Attribution

Include the exact Java file from JTS that each ported Go file was derived from
(this is implied by the source file name, but it would be nice to make this
explicit).
