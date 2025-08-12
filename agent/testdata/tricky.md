   # Three leading spaces: valid heading

    # Four leading spaces: this is a code block line, NOT a heading

Text line followed by underline with too much indent:
    ---
This should not be considered a setext heading because of 4 spaces.

A setext heading with a blank line between:
Title

---
This might be treated as a heading by our function.
