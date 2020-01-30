"""
The version package contains warning messages that should be displayed
to a Go plugin user when upgrading from an older version of the plugin
"""

import semver

"""
If there are breaking changes that need to be communicated to a user,
add them to this dict. For readability, an opening and closing newline
is recommended for each warning message
"""
WARNINGS = {
    # FIXME: Version number to be finalised
    semver.VersionInfo(
        0, 1, 3
    ): """
Generated models no longer use the types exported in the encoding package.
Your model's fields have been regenerated using standard pointer types (*string, *int, etc) as used in the AWS Go SDK.
The AWS SDK has helper functions that you can use to get and set your model's values.

Make the following changes to your handler code as needed:

* Replace `encoding.New{Type}` with `aws.{Type}`
* Replace `model.{field}.Value()` with `aws.{Type}Value(model.{field})`

Where {Type} is either String, Bool, Int, or Float64 and {field} is any field within your generated model.
""",
}


def check_version(current_version):
    """
    check_version compares the user's current plugin version with each
    version in WARNINGS and returns any appropriate messages
    """

    if current_version is not None:
        current_version = semver.VersionInfo.parse(current_version)

    return [
        f"Change message for Go plugin v{version}:" + WARNINGS[version]
        for version in sorted(WARNINGS.keys())
        if current_version is None or current_version < version
    ]
