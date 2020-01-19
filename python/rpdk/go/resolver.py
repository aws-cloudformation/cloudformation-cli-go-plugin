from rpdk.core.jsonutils.resolver import UNDEFINED, ContainerType

PRIMITIVE_TYPES = {
    "string": "*string",
    "integer": "*int",
    "boolean": "*bool",
    "number": "*float",
    UNDEFINED: "interface{}",
}


def translate_type(resolved_type):
    if resolved_type.container == ContainerType.MODEL:
        return "*" + resolved_type.type
    if resolved_type.container == ContainerType.PRIMITIVE:
        return PRIMITIVE_TYPES[resolved_type.type]

    item_type = translate_type(resolved_type.type)

    if resolved_type.container == ContainerType.DICT:
        return f"map[string]{item_type}"
    if resolved_type.container == ContainerType.LIST:
        return f"[]{item_type}"
    if resolved_type.container == ContainerType.SET:
        return f"[]{item_type}"

    raise ValueError(f"Unknown container type {resolved_type.container}")
