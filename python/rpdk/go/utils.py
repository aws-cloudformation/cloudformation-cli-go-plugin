from typing import Callable

# https://golang.org/ref/spec#Keywords
LANGUAGE_KEYWORDS = {
    "break",
    "default",
    "func",
    "interface",
    "select",
    "case",
    "defer",
    "go",
    "map",
    "struct",
    "chan",
    "else",
    "goto",
    "package",
    "switch",
    "const",
    "fallthrough",
    "if",
    "range",
    "type",
    "continue",
    "for",
    "import",
    "return",
    "var",
}


def safe_reserved(string: str) -> str:
    if string in LANGUAGE_KEYWORDS:
        return string + "_"
    return string


def validate_path(default: str) -> Callable[[str], str]:
    def _validate_namespace(value: str) -> str:
        if not value:
            return default

        namespace = value

        return namespace

    return _validate_namespace
