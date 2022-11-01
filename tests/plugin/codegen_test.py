# pylint: disable=redefined-outer-name,protected-access
from pathlib import Path
from shutil import copyfile
from unittest.mock import patch

import pytest

from rpdk.core.exceptions import DownstreamError
from rpdk.core.project import Project
from rpdk.go.__init__ import __version__
from rpdk.go.codegen import GoLanguagePlugin

TYPE_NAME = "foo::bar::baz"

TEST_TARGET_INFO = {
    "My::Example::Resource": {
        "TargetName": "My::Example::Resource",
        "TargetType": "RESOURCE",
        "Schema": {
            "typeName": "My::Example::Resource",
            "additionalProperties": False,
            "properties": {
                "Id": {"type": "string"},
                "Tags": {
                    "type": "array",
                    "uniqueItems": False,
                    "items": {"$ref": "#/definitions/Tag"},
                },
            },
            "required": [],
            "definitions": {
                "Tag": {
                    "type": "object",
                    "additionalProperties": False,
                    "properties": {
                        "Value": {"type": "string"},
                        "Key": {"type": "string"},
                    },
                    "required": ["Value", "Key"],
                }
            },
        },
        "ProvisioningType": "FULLY_MUTTABLE",
        "IsCfnRegistrySupportedType": True,
        "SchemaFileAvailable": True,
    },
    "My::Other::Resource": {
        "TargetName": "My::Other::Resource",
        "TargetType": "RESOURCE",
        "Schema": {
            "typeName": "My::Other::Resource",
            "additionalProperties": False,
            "properties": {
                "Id": {"type": "string"},
                "Tags": {
                    "type": "array",
                    "uniqueItems": False,
                    "items": {"$ref": "#/definitions/Tag"},
                },
            },
            "required": [],
            "definitions": {
                "Tag": {
                    "type": "object",
                    "additionalProperties": False,
                    "properties": {
                        "Value": {"type": "string"},
                        "Key": {"type": "string"},
                    },
                    "required": ["Value", "Key"],
                }
            },
        },
        "ProvisioningType": "NOT_PROVISIONABLE",
        "IsCfnRegistrySupportedType": False,
        "SchemaFileAvailable": True,
    },
}

@pytest.fixture
def plugin():
    return GoLanguagePlugin()


@pytest.fixture
def resource_project(tmp_path):
    project = Project(root=tmp_path)

    patch_plugins = patch.dict(
        "rpdk.core.plugin_registry.PLUGIN_REGISTRY",
        {GoLanguagePlugin.NAME: lambda: GoLanguagePlugin},
        clear=True,
    )
    patch_wizard = patch(
        "rpdk.go.codegen.input_with_validation", autospec=True, side_effect=[False]
    )
    with patch_plugins, patch_wizard:
        project.init(TYPE_NAME, GoLanguagePlugin.NAME)
    return project


def get_files_in_project(project):
    return {
        str(child.relative_to(project.root)): child for child in project.root.rglob("*")
    }

def test_initialize_resource(resource_project):
    assert resource_project.settings == {
        "import_path": 'False',
        "protocolVersion": "2.0.0",
        "pluginVersion": "2.0.4",
        "use_docker": None,
    }

    files = get_files_in_project(resource_project)
    assert set(files) == {
        ".gitignore",
        ".rpdk-config",
        "Makefile",
        "README.md",
        "cmd",
        "cmd/resource",
        "cmd/resource/resource.go",
        "foo-bar-baz.json",
        "go.mod",
        "internal",
        "example_inputs/inputs_1_invalid.json",
        "example_inputs/inputs_1_update.json",
        "example_inputs/inputs_1_create.json",
        "example_inputs",
        "template.yml",
    }

    readme = files["README.md"].read_text()
    assert resource_project.type_name in readme

    assert resource_project.entrypoint in files["template.yml"].read_text()

def test_generate_resource(resource_project):
    resource_project.load_schema()
    before = get_files_in_project(resource_project)
    resource_project.generate()
    after = get_files_in_project(resource_project)
    files = after.keys() - before.keys() - {"resource-role.yaml"}

    assert files == {
        "makebuild",
        "cmd/main.go",
        "cmd/resource/config.go",
        "cmd/resource/model.go",
    }


def test_generate_resource_go_failure(resource_project):
    resource_project.load_schema()

    with patch('rpdk.go.codegen.subprocess_run') as mock_subprocess:
        mock_subprocess.side_effect = FileNotFoundError()
        with pytest.raises(DownstreamError, match='go fmt failed'):
            resource_project.generate()


def test_generate_resource_with_type_configuration(tmp_path):
    type_name = "schema::with::typeconfiguration"
    project = Project(root=tmp_path)

    patch_plugins = patch.dict(
        "rpdk.core.plugin_registry.PLUGIN_REGISTRY",
        {GoLanguagePlugin.NAME: lambda: GoLanguagePlugin},
        clear=True,
    )
    patch_wizard = patch(
        "rpdk.go.codegen.input_with_validation", autospec=True, side_effect=[False]
    )
    with patch_plugins, patch_wizard:
        project.init(type_name, GoLanguagePlugin.NAME)

    copyfile(
        str(Path.cwd() / "tests/data/schema-with-typeconfiguration.json"),
        str(project.root / "schema-with-typeconfiguration.json"),
    )
    project.type_info = ("schema", "with", "typeconfiguration")
    project.load_schema()
    project.load_configuration_schema()
    project.generate()

    type_configuration_schema_file = (
        project.root / "schema-with-typeconfiguration-configuration.json"
    )
    assert type_configuration_schema_file.is_file()
