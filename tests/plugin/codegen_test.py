# pylint: disable=redefined-outer-name,protected-access
import ast
import importlib.util
from pathlib import Path
from shutil import copyfile
from subprocess import CalledProcessError
from unittest.mock import ANY, patch, sentinel
from uuid import uuid4
from zipfile import ZipFile

import pytest
from docker.errors import APIError, ContainerError, ImageLoadError
from requests.exceptions import ConnectionError as RequestsConnectionError
from rpdk.core.exceptions import DownstreamError
from rpdk.core.project import Project
from rpdk.go.codegen import GoLanguagePlugin

TYPE_NAME = "foo::bar::baz"


@pytest.fixture
def plugin():
    return GoLanguagePlugin()


@pytest.fixture
def project(tmp_path):
    project = Project(root=tmp_path)
    patch_plugins = patch.dict(
        "rpdk.core.plugin_registry.PLUGIN_REGISTRY",
        {GoLanguagePlugin.RUNTIME: lambda: GoLanguagePlugin},
        clear=True,
    )
    patch_wizard = patch(
        "rpdk.go.codegen.input_with_validation", autospec=True, side_effect=[False]
    )
    with patch_plugins, patch_wizard:
        project.init(TYPE_NAME, GoLanguagePlugin.RUNTIME)
    return project


def get_files_in_project(project):
    return {
        str(child.relative_to(project.root)): child for child in project.root.rglob("*")
    }


def test_initialize(project):
    print(project.settings)
    assert project.settings == {
        "import_path": "False",
        "protocolVersion": "2.0.0",
        "pluginVersion": "2.0.1",
    }
    files = get_files_in_project(project)
    assert set(files) == {
        ".gitignore",
        ".rpdk-config",
        "README.md",
        "foo-bar-baz.json",
        "example_inputs/inputs_1_invalid.json",
        "example_inputs/inputs_1_update.json",
        "example_inputs/inputs_1_create.json",
        "example_inputs",
        "cmd",
        "cmd/resource/resource.go",
        "template.yml",
        "cmd/resource",
        "go.mod",
        "internal",
        "Makefile",
    }


def test_generate(project):
    project.load_schema()
    before = get_files_in_project(project)
    project.generate()
    after = get_files_in_project(project)
    files = (
        after.keys()
        - before.keys()
        - {"resource-role.yaml", "cmd/main.go", "makebuild"}
    )
    assert files == {"cmd/resource/model.go", "go.sum"}
    type_configuration_schema_file = project.root / "foo-bar-baz-configuration.json"
    assert not type_configuration_schema_file.is_file()


def test_generate_with_type_configuration(tmp_path):
    type_name = "schema::with::typeconfiguration"
    project = Project(root=tmp_path)

    patch_plugins = patch.dict(
        "rpdk.core.plugin_registry.PLUGIN_REGISTRY",
        {GoLanguagePlugin.RUNTIME: lambda: GoLanguagePlugin},
        clear=True,
    )
    patch_wizard = patch(
        "rpdk.go.codegen.input_with_validation", autospec=True, side_effect=[False]
    )
    with patch_plugins, patch_wizard:
        project.init(type_name, GoLanguagePlugin.RUNTIME)

    copyfile(
        str(Path.cwd() / "data/schema-with-typeconfiguration.json"),
        str(project.root / "schema-with-typeconfiguration.json"),
    )
    project.type_info = ("schema", "with", "typeconfiguration")
    project.load_schema()
    project.load_configuration_schema()
    project.generate()

    # assert TypeConfigurationModel is added to generated directory
    # models_path = project.root / "src" / "schema_with_typeconfiguration" / "models.py"

    # this however loads the module
    # spec = importlib.util.spec_from_file_location("foo_bar_baz.models", models_path)
    # module = importlib.util.module_from_spec(spec)
    # spec.loader.exec_module(module)

    # assert hasattr(module.ResourceModel, "_serialize")
    # assert hasattr(module.ResourceModel, "_deserialize")
    # assert hasattr(module.TypeConfigurationModel, "_serialize")
    # assert hasattr(module.TypeConfigurationModel, "_deserialize")

    type_configuration_schema_file = (
        project.root / "schema-with-typeconfiguration-configuration.json"
    )
    assert type_configuration_schema_file.is_file()
