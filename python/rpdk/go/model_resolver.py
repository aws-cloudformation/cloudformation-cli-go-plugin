from rpdk.core.filters import uppercase_first_letter
from rpdk.core.jsonutils.utils import BASE, fragment_encode


class ModelResolverError(Exception):
    pass


class CSharpModelResolver:
    """This class takes in a flattened schema map (output of the JsonSchemaFlattener),
    and builds a full set of Go types.
    """

    def __init__(self, flattened_schema_map, resource_type):
        self.flattened_schema_map = flattened_schema_map
        self._resource_class_name = uppercase_first_letter(resource_type)
        self._ref_to_class_map = self._get_ref_to_class_map()

    def _get_ref_to_class_map(self):
        """Creates a Go type name for each ref_path in the flattened schema map.
        """
        ref_to_class_map = {(): self.resource_class_name}
        for ref_path in self.flattened_schema_map.keys():
            if ref_path == ():
                continue
            ref_to_class_map[ref_path] = self._get_class_name_from_ref(
                ref_path, ref_to_class_map
            )
        return ref_to_class_map

    @staticmethod
    def _get_class_name_from_ref(ref_path, ref_to_class_map):
        """Given a json schema ref, returns the best guess at a Go type name.
        """
        class_name = base_class_from_ref(ref_path)

        # TODO: resolve duplicate class names using subfolders
        while class_name in ref_to_class_map.values():
            class_name += "_"
        return class_name

    @property
    def resource_class_name(self):
        return self._resource_class_name

    def resolve_models(self):
        """Main method of the class that iterates through each schema and creates
        the Go type map.

        :return: a map where the keys are Go type names, and the values are a map
        of the defined property names to Go property types.
        """
        models = {}
        for ref_path, sub_schema in self.flattened_schema_map.items():
            class_name = self._ref_to_class_map[ref_path]
            model_property_map = {}
            for prop_name, prop_schema in sub_schema["properties"].items():
                model_property_map[prop_name] = self._csharp_property_type(prop_schema)
            models[class_name] = model_property_map
        return models

    def _csharp_property_type(self, property_schema):
        """Return the Go type for a flattened schema.
        If the schema is a ref, the class is determined from the ref_to_class_map
        """
        try:
            ref_path = property_schema["$ref"]
        except KeyError:
            pass  # we are not dealing with a ref, move on
        else:
            return self._ref_to_class_map[ref_path]

        json_type = property_schema.get("type", "object")

        if json_type == "array":
            return self._csharp_array_type(property_schema)

        if json_type == "object":
            return self._csharp_object_type(property_schema)

        primitive_types_map = {
            "string": "string",
            "integer": "int",
            "boolean": "bool",
            "number": "float64",
        }

        return primitive_types_map[json_type]

    def _csharp_array_type(self, property_schema):
        """For an array type, we create a Go slice.
        """
        try:
            items = property_schema["items"]
        except KeyError:
            array_items_class_name = "interface{}"
        else:
            array_items_class_name = self._csharp_property_type(items)

        return "[]{}".format(array_items_class_name)

    def _csharp_object_type(self, property_schema):
        """Resolves an array type schema to a C# class.  An object type will
        always be a Dictionary<String, V>
        * If patternProperties is defined, V is determined by the schema for the
        pattern. We do not care about the pattern itself, since that is only used
        for validation.
        * The object will never have nested properties, as that was taken care of by
        flattening the schema
        * If there are no patternProperties, it must be an arbitrary JSON type, so V
        will be an object.
        """
        try:
            pattern_properties = list(property_schema["patternProperties"].items())
        except KeyError:
            return "Dictionary<string, object>"  # no pattern properties == object type
        else:
            if len(pattern_properties) != 1:
                return "Dictionary<string, object>"  # bad schema definition
            pattern_properties_class_name = self._csharp_property_type(
                pattern_properties[0][1]
            )
            return "Dictionary<string, {}>".format(pattern_properties_class_name)



def base_class_from_ref(ref_path):
    """This method determines the class_name from a ref_path
    It uses json-schema heuristics to properly determine the class name

    >>> base_class_from_ref(("definitions", "Foo"))
    'Foo'
    >>> base_class_from_ref(("properties", "foo", "items"))
    'Foo'
    >>> base_class_from_ref(("properties", "foo", "items", "patternProperties", "a"))
    'Foo'
    >>> base_class_from_ref(("properties", "items"))
    'Items'
    >>> base_class_from_ref(("properties", "patternProperties"))
    'PatternProperties'
    >>> base_class_from_ref(("properties", "properties"))
    'Properties'
    >>> base_class_from_ref(("definitions",))
    'Definitions'
    >>> base_class_from_ref(("definitions", "properties"))
    'Properties'
    >>> base_class_from_ref(())   # doctest: +NORMALIZE_WHITESPACE
    Traceback (most recent call last):
    ...
    csharp.model_resolver.ModelResolverError:
    Could not create a valid class from schema at '#'
    """
    parent_keywords = ("properties", "definitions")
    schema_keywords = ("items", "patternProperties", "properties")

    ref_parts = ref_path[::-1]
    ref_parts_with_root = ref_parts + (BASE,)
    for idx, elem in enumerate(ref_parts):
        parent = ref_parts_with_root[idx + 1]
        if parent in parent_keywords or (
            elem not in schema_keywords and parent != "patternProperties"
        ):
            return uppercase_first_letter(elem.rpartition("/")[2])

    raise ModelResolverError(
        "Could not create a valid class from schema at '{}'".format(
            fragment_encode(ref_path)
        )
    )
