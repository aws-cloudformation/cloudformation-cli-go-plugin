from rpdk.core.jsonutils.resolver import UNDEFINED, ContainerType

PRIMITIVE_TYPES = {
    "string": "string",
    "integer": "int",
    "boolean": "bool",
    "number": "float64",
    UNDEFINED: "interface{}",
}


def translate_item_type(resolved_type):
    """
    translate_item_type converts JSON schema item types into Go types
    """

    # Another model
    if resolved_type.container == ContainerType.MODEL:
        return resolved_type.type

    # Primitive type
    if resolved_type.container == ContainerType.PRIMITIVE:
        return PRIMITIVE_TYPES[resolved_type.type]

    # Something more complex
    return translate_type(resolved_type)


def translate_type(resolved_type):
    """
    translate_type converts JSON schema types into Go types
    """

    # Another model
    if resolved_type.container == ContainerType.MODEL:
        return "*" + resolved_type.type

    # Primitive type
    if resolved_type.container == ContainerType.PRIMITIVE:
        return "*" + PRIMITIVE_TYPES[resolved_type.type]

    # Composite type
    item_type = translate_item_type(resolved_type.type)

    # A dict
    if resolved_type.container == ContainerType.DICT:
        return f"map[string]{item_type}"

    # A list
    if resolved_type.container in (ContainerType.LIST, ContainerType.SET):
        return f"[]{item_type}"

    raise ValueError(f"Unknown container type {resolved_type.container}")
