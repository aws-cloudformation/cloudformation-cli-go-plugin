# pylint: disable=useless-super-delegation,too-many-locals
# pylint doesn't recognize abstract methods
import logging
import shutil
import os

from rpdk.core.data_loaders import resource_stream
from rpdk.core.jsonutils.flattener import JsonSchemaFlattener
from rpdk.core.plugin_base import LanguagePlugin

from .model_resolver import CSharpModelResolver
from .utils import safe_reserved
from subprocess import call

LOG = logging.getLogger(__name__)

OPERATIONS = ("create", "read", "update", "delete", "list")
EXECUTABLE = "cfn-cli"


class GoLanguagePlugin(LanguagePlugin):
    MODULE_NAME = __name__
    RUNTIME = "go1.x"
    ENTRY_POINT = "handler"
    CODE_URI = "./bin"

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
        project.runtime = self.RUNTIME
        project.entrypoint = self.ENTRY_POINT.format(self.package_name)

        # .gitignore
        path = project.root / ".gitignore"
        LOG.debug("Writing .gitignore: %s", path)
        contents = resource_stream(__name__, "data/go.gitignore").read()
        project.safewrite(path, contents)

        # project folder structure
        src = (project.root / "cmd"  / "resource")
        LOG.debug("Making source folder structure: %s", src)
        src.mkdir(parents=True, exist_ok=True)

        inter = (project.root / "internal")
        inter.mkdir(parents=True, exist_ok=True)


        # Makefile    
        path = project.root / "Makefile"
        LOG.debug("Writing Makefile: %s", path)
        template = self.env.get_template("Makefile")
        contents = template.render()
        project.safewrite(path, contents)
        
        # CloudFormation/SAM template for handler lambda
        path = project.root / "template.yml"
        LOG.debug("Writing SAM template: %s", path)
        template = self.env.get_template("template.yml")

        handler_params = {
            "Handler": project.entrypoint,
            "Runtime": project.runtime,
            "CodeUri": self.CODE_URI,
        }
        contents = template.render(
            resource_type=project.type_name,
            functions={
                "TypeFunction": handler_params,
                "TestEntrypoint": {
                    **handler_params,
                    "Handler": handler_params["Handler"].replace(
                        "handleRequest", "testEntrypoint"
                    ),
                },
            },
        )
        project.safewrite(path, contents)

        LOG.debug("Writing handlers and tests")
        self.init_handlers(project, src)

        template = self.env.get_template("callback.go.tple")
        LOG.debug("Writing Callback Context")
        path = src / "{}.go".format("callbackContext")
        contents = template.render()
        project.safewrite(path, contents)

        # README
        path = project.root / "README.md"
        LOG.debug("Writing README: %s", path)
        template = self.env.get_template("README.md")
        contents = template.render(
            type_name=project.type_name,
            schema_path=project.schema_path,
            executable=EXECUTABLE,
            files="generated.go and main.go"
        )
        project.safewrite(path, contents)
        
        LOG.debug("Init complete")

    def init_handlers(self, project, src):
        LOG.debug("Writing stub handlers")
        for operation in OPERATIONS:
            if operation == "list":
                template = self.env.get_template("StubListHandler.go.tple")
            else:
                template = self.env.get_template("StubHandler.go.tple")
            path = src / "{}Handler.go".format(operation)
            LOG.debug("%s handler: %s", operation, path)
            contents = template.render(
                model_name=self.namespace[2],
                operation=operation,
            )
            project.safewrite(path, contents)

    def _get_generated_root(self, project):
        self._namespace_from_project(project)
        return project.root / "cmd"  /  "resource"

    def generate(self, project):
        LOG.debug("Generate started")
   
        self._namespace_from_project(project)

        objects = JsonSchemaFlattener(project.schema).flatten_schema()

        # project folder structure
        src = (project.root / "cmd"  / "resource")
        
        model_resolver = CSharpModelResolver(objects, "Resource")
        models = model_resolver.resolve_models()

        LOG.debug("Writing %d models", len(models))

        
        template = self.env.get_template("model.go.tple")
        for model_name, properties in models.items():
            path = src / "{}.go".format("generated")
            LOG.debug("%s model: %s", model_name, path)
            contents = template.render(
                package_name=self.package_name,
                model_name=self.namespace[2].capitalize(),
                properties=properties,
            )
            project.overwrite(path, contents)
        
        path = project.root / "cmd"  / "main.go"
        parts = os.path.split(path)
        LOG.debug("Writing project: %s", path)
        template = self.env.get_template("main.go.tple")
        gopath='{}/src/'.format(os.environ['GOPATH'])
        parts = parts[0].split(gopath)
        contents = template.render(
            model_name=self.namespace[2],
            path=parts[1] + '/resource'
        )
        project.overwrite(path, contents)
        
        
        LOG.debug("Generate complete")
       
    def package(self, project):
        pass
        
