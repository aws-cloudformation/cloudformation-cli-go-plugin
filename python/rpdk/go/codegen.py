# pylint: disable=useless-super-delegation,too-many-locals
# pylint doesn't recognize abstract methods
import logging
import shutil
import os

from rpdk.core.jsonutils.flattener import JsonSchemaFlattener
from rpdk.core.plugin_base import LanguagePlugin

from .model_resolver import CSharpModelResolver
from .utils import safe_reserved

LOG = logging.getLogger(__name__)

OPERATIONS = ("Create", "Read", "Update", "Delete", "List")
EXECUTABLE = "uluru-cli"


class GoLanguagePlugin(LanguagePlugin):
    MODULE_NAME = __name__
    NAME = "golang"
    RUNTIME = "dotnetcore2.0"
    ENTRY_POINT = "{}.LambdaInterceptor::InterceptRequest"
    CODE_URI = "./bin/Release/netstandard2.0/ResourceProvider.dll"

    def __init__(self):
        self.env = self._setup_jinja_env(
            trim_blocks=True, lstrip_blocks=True, keep_trailing_newline=True
        )
        self.namespace = None
        self.package_name = None

    def _namespace_from_project(self, project):
        self.namespace = tuple(
            safe_reserved(s.title()) for s in project.type_info
        )
        self.package_name = ".".join(self.namespace)

    def init(self, project):
        LOG.debug("Init started")

        self._namespace_from_project(project)

        # project folder structure
        src = (project.root / "cmd"  / self.namespace[2].lower() / "resource")
        LOG.debug("Making source folder structure: %s", src)
        src.mkdir(parents=True, exist_ok=True)
        
        tst = (project.root / "cmd"  / self.namespace[2].lower() / "test")
        LOG.debug("Making test folder structure: %s", tst)
        tst.mkdir(parents=True, exist_ok=True)
        
        
        path = project.root / "cmd"  / self.namespace[2].lower() / "main.go"
        LOG.debug("Writing project: %s", path)
        template = self.env.get_template("main.go")
        contents = template.render(
            model_name=self.namespace[2],
            path=project.root / "cmd"  / self.namespace[2].lower() / "resource",
        )
        project.safewrite(path, contents)

        
        path = project.root / "Makefile"
        LOG.debug("Writing Makefile: %s", path)
        template = self.env.get_template("Makefile")
        contents = template.render(
            model_name=self.namespace[2].lower(),
        )
        project.safewrite(path, contents)
        
        # CloudFormation/SAM template for handler lambda
        path = project.root / "Handler.yaml"
        LOG.debug("Writing SAM template: %s", path)
        template = self.env.get_template("Handler.yaml")
        contents = template.render(
            resource_type=project.type_name,
        )
        project.safewrite(path, contents)
        '''
        LOG.debug("Writing stub handlers")
        template = self.env.get_template("StubHandler.cs")

        for operation in OPERATIONS:
            path = src / "{}Handler.cs".format(operation)
            LOG.debug("%s handler: %s", operation, path)
            contents = template.render(
                package_name=self.package_name,
                operation=operation,
                model_name="ResourceModel",
            )
            project.safewrite(path, contents)

        path = src / "Configuration.cs"
        LOG.debug("Writing configuration: %s", path)
        template = self.env.get_template("StubConfiguration.cs")
        contents = template.render(
            package_name=self.package_name, schema_file_name=project.schema_filename
        )
        project.safewrite(path, contents)

        path = project.root / "README.md"
        LOG.debug("Writing README: %s", path)eneccccbiekdcjvngjnceegtbeibtbjbgdbjkvuvctre

        template = self.env.get_template("README.md")
        contents = template.render(
            type_name=project.type_name,
            schema_path=project.schema_path,
            executable=EXECUTABLE,
            generated_root=self._get_generated_root(project)
        )
        project.safewrite(path, contents)
      '''
        LOG.debug("Init complete")

    
    def _get_generated_root(self, project):
        self._namespace_from_project(project)
        return project.root / "cmd"  / self.namespace[2].lower() / "resource"

    def generate(self, project):
        LOG.debug("Generate started")

        
        self._namespace_from_project(project)

        objects = JsonSchemaFlattener(project.schema).flatten_schema()

        generated_root = self._get_generated_root(project)
        LOG.debug("Removing generated sources: %s", generated_root)
        shutil.rmtree(generated_root, ignore_errors=True)
        
        # project folder structure
        src = (project.root / "cmd"  / self.namespace[2].lower() / "resource")
        LOG.debug("Making resource folder structure: %s", src)
        src.mkdir(parents=True, exist_ok=True)

        model_resolver = CSharpModelResolver(objects, "Resource")
        models = model_resolver.resolve_models()

        LOG.debug("Writing %d models", len(models))

        template = self.env.get_template("Model.go")
        for model_name, properties in models.items():
            path = src / "{}.go".format(model_name)
            LOG.debug("%s model: %s", model_name, path)
            contents = template.render(
                package_name=self.package_name,
                model_name=model_name,
                properties=properties,
            )
            project.overwrite(path, contents)
        LOG.debug("Format type")
        myCmd = "gofmt -w {}".format(path)
        os.system(myCmd)
        
        '''
        path = src / "HandlerWrapper.cs"
        LOG.debug("Writing handler wrapper: %s", path)
        template = self.env.get_template("HandlerWrapper.cs")
        contents = template.render(
            package_name=self.package_name,
            operations=OPERATIONS,
            model_name="ResourceModel",
        )
        project.overwrite(path, contents)

        path = src / "BaseConfiguration.cs"
        LOG.debug("Writing base configuration: %s", path)
        template = self.env.get_template("BaseConfiguration.cs")
        contents = template.render(
            package_name=self.package_name, schema_file_name=project.schema_filename
        )
        project.overwrite(path, contents)

        path = src / "BaseHandler.cs"
        LOG.debug("Writing base handler: %s", path)
        template = self.env.get_template("BaseHandler.cs")
        contents = template.render(
            package_name=self.package_name,
            operations=OPERATIONS,
            model_name="ResourceModel",
        )
        project.overwrite(path, contents)

        model_resolver = CSharpModelResolver(objects, "ResourceModel")
        models = model_resolver.resolve_models()

        LOG.debug("Writing %d models", len(models))

        template = self.env.get_template("Model.cs")
        for model_name, properties in models.items():
            path = src / "{}.cs".format(model_name)
            LOG.debug("%s model: %s", model_name, path)
            contents = template.render(
                package_name=self.package_name,
                model_name=model_name,
                properties=properties,
            )
            project.overwrite(path, contents)
        '''
        LOG.debug("Generate complete")

    def package(self, project):
        pass
