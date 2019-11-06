from rpdk.core.jsonutils.resolver import UNDEFINED, ContainerType

PRIMITIVE_TYPES = {
    "string": "*encoding.String",
    "integer": "*encoding.Int",
    "boolean": "*encoding.Bool",
    "number": "*encoding.Float",
    UNDEFINED: "Type",
}


def translate_type(resolved_type):
    if resolved_type.container == ContainerType.MODEL:
        return resolved_type.type
    if resolved_type.container == ContainerType.PRIMITIVE:
        return PRIMITIVE_TYPES[resolved_type.type]

    item_type = translate_type(resolved_type.type)

    if resolved_type.container == ContainerType.DICT:
        key_type = PRIMITIVE_TYPES["string"]
        return f"Map<{key_type}, {item_type}>"
    if resolved_type.container == ContainerType.LIST:
        return f"[]{item_type}"
    if resolved_type.container == ContainerType.SET:
        return f"[]{item_type}"

    raise ValueError(f"Unknown container type {resolved_type.container}")
