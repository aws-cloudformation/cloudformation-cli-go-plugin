# pylint: disable=useless-super-delegation,too-many-locals
# pylint doesn't recognize abstract methods
import logging
import zipfile
from pathlib import Path
from rpdk.core.data_loaders import resource_stream
from rpdk.core.exceptions import DownstreamError, InternalError, SysExitRecommendedError
from rpdk.core.init import input_with_validation
from rpdk.core.jsonutils.resolver import resolve_models
from rpdk.core.plugin_base import LanguagePlugin
from rpdk.core.project import Project
from subprocess import PIPE, CalledProcessError, run as subprocess_run  # nosec
from tempfile import TemporaryFile

from .resolver import translate_type
from .utils import safe_reserved, validate_path

LOG = logging.getLogger(__name__)

OPERATIONS = ("Create", "Read", "Update", "Delete", "List")
EXECUTABLE = "cfn-cli"

LANGUAGE = "go"

DEFAULT_SETTINGS = {"protocolVersion": "2.0.0"}


class GoExecutableNotFoundError(SysExitRecommendedError):
    pass


class GoLanguagePlugin(LanguagePlugin):
    MODULE_NAME = __name__
    NAME = "go"
    RUNTIME = "provided.al2"
    ENTRY_POINT = "bootstrap"
    TEST_ENTRY_POINT = "bootstrap"
    CODE_URI = "bin/"

    def __init__(self):
        self.env = self._setup_jinja_env(
            trim_blocks=True, lstrip_blocks=True, keep_trailing_newline=True
        )
        self.env.filters["translate_type"] = translate_type
        self.env.filters["safe_reserved"] = safe_reserved
        self._use_docker = None
        self._protocol_version = "2.0.0"
        self.import_path = ""

    def _prompt_for_go_path(self, project):
        path_validator = validate_path("")
        import_path = path_validator(project.settings.get("import_path"))

        if not import_path:
            prompt = "Enter the GO Import path"
            import_path = input_with_validation(prompt, path_validator)

        self.import_path = import_path
        project.settings["import_path"] = str(self.import_path)

    def init(self, project: Project):
        LOG.debug("Init started")

        self._prompt_for_go_path(project)

        self._init_settings(project)

        # .gitignore
        path = project.root / ".gitignore"
        LOG.debug("Writing .gitignore: %s", path)
        contents = resource_stream(__name__, "data/go.gitignore").read()
        project.safewrite(path, contents)

        # project folder structure
        src = project.root / "cmd" / "resource"
        LOG.debug("Making source folder structure: %s", src)
        src.mkdir(parents=True, exist_ok=True)

        inter = project.root / "internal"
        inter.mkdir(parents=True, exist_ok=True)

        # Makefile
        path = project.root / "Makefile"
        LOG.debug("Writing Makefile: %s", path)
        template = self.env.get_template("Makefile")
        contents = template.render()
        project.overwrite(path, contents)

        # go.mod
        path = project.root / "go.mod"
        LOG.debug("Writing go.mod: %s", path)
        template = self.env.get_template("go.mod.tple")
        contents = template.render(path=Path(project.settings["import_path"]))
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
        test_handler_params = {
            "Handler": project.entrypoint,
            "Runtime": project.runtime,
            "CodeUri": self.CODE_URI,
            "Environment": "",
            "  Variables": "",
            "    MODE": "Test",
        }
        contents = template.render(
            resource_type=project.type_name,
            functions={
                "TypeFunction": handler_params,
                "TestEntrypoint": test_handler_params,
            },
        )
        project.safewrite(path, contents)

        LOG.debug("Writing handlers and tests")
        self.init_handlers(project, src)

        # README
        path = project.root / "README.md"
        LOG.debug("Writing README: %s", path)
        template = self.env.get_template("README.md")
        contents = template.render(
            type_name=project.type_name,
            schema_path=project.schema_path,
            executable=EXECUTABLE,
            files="model.go and main.go",
        )
        project.safewrite(path, contents)

        LOG.debug("Init complete")

    def _init_settings(self, project: Project):
        project.runtime = self.RUNTIME
        project.entrypoint = self.ENTRY_POINT.format(self.import_path)
        project.test_entrypoint = self.TEST_ENTRY_POINT.format(self.import_path)
        project.settings.update(DEFAULT_SETTINGS)
        if project.settings.get("use_docker"):
            self._use_docker = True
        else:
            self._use_docker = False
        project.settings["protocolVersion"] = self._protocol_version

    def init_handlers(self, project: Project, src):
        LOG.debug("Writing stub handlers")
        template = self.env.get_template("stubHandler.go.tple")
        path = src / "resource.go"
        contents = template.render()
        project.safewrite(path, contents)

    # pylint: disable=unused-argument
    def _get_generated_root(self, project: Project):
        LOG.debug("Init started")

    def generate(self, project: Project):
        LOG.debug("Generate started")
        root = project.root / "cmd"

        # project folder structure
        src = root / "resource"
        format_paths = []

        LOG.debug("Writing Types")

        models = resolve_models(project.schema)
        if project.configuration_schema:
            configuration_schema_path = (
                project.root / project.configuration_schema_filename
            )
            project.write_configuration_schema(configuration_schema_path)
            configuration_models = resolve_models(
                project.configuration_schema, "TypeConfiguration"
            )
        else:
            configuration_models = {"TypeConfiguration": {}}

        # Create the type configuration model
        template = self.env.get_template("config.go.tple")
        path = src / "config.go"
        contents = template.render(models=configuration_models)
        project.overwrite(path, contents)
        format_paths.append(path)

        # Create the resource model
        template = self.env.get_template("types.go.tple")
        path = src / "model.go"
        contents = template.render(models=models)
        project.overwrite(path, contents)
        format_paths.append(path)

        path = root / "main.go"
        LOG.debug("Writing project: %s", path)
        template = self.env.get_template("main.go.tple")
        importpath = Path(project.settings["import_path"])
        contents = template.render(path=(importpath / "cmd" / "resource").as_posix())
        project.overwrite(path, contents)
        format_paths.append(path)

        # makebuild
        path = project.root / "makebuild"
        LOG.debug("Writing makebuild: %s", path)
        template = self.env.get_template("makebuild")
        contents = template.render()
        project.overwrite(path, contents)

        # named files must all be in one directory
        for path in format_paths:
            try:
                subprocess_run(
                    ["go", "fmt", path], cwd=root, check=True, stdout=PIPE, stderr=PIPE
                )  # nosec
            except (FileNotFoundError, CalledProcessError) as e:
                raise DownstreamError("go fmt failed") from e

        # Update settings as needed
        need_to_write = False
        for key, new in DEFAULT_SETTINGS.items():
            old = project.settings.get(key)

            if project.settings.get(key) != new:
                LOG.debug(
                    "{key} version change from {old} to {new}",
                    key=key,
                    old=old,
                    new=new,
                )
                project.settings[key] = new
                need_to_write = True

        if need_to_write:
            project.write_settings()

    @staticmethod
    def pre_package(project: Project):
        # zip the Go build output - it's all needed to execute correctly
        f = TemporaryFile("w+b")  # pylint: disable=R1732

        with zipfile.ZipFile(f, mode="w") as zip_file:
            for path in (project.root / "bin").iterdir():
                if path.is_file():
                    zip_file.write(path.resolve(), path.name)
        f.seek(0)

        return f

    @staticmethod
    def _find_exe(project: Project):
        exe_glob = list((project.root / "bin").glob("bootstrap"))
        if not exe_glob:
            LOG.debug("No Go executable match")
            raise GoExecutableNotFoundError(
                "You must build the handler before running cfn-submit.\n"
                "Please run 'make' or the equivalent command "
                "in your IDE to compile and package the code."
            )

        if len(exe_glob) > 1:
            LOG.debug(
                "Multiple Go executable match: %s",
                ", ".join(str(path) for path in exe_glob),
            )
            raise InternalError("Multiple Go executable match")

        LOG.debug("Generate complete")
        return exe_glob[0]

    def package(self, project: Project, zip_file):
        LOG.info("Packaging Go project")

        def write_with_relative_path(path):
            relative = path.relative_to(project.root)
            zip_file.write(path.resolve(), str(relative))

        # sanity check for build output
        self._find_exe(project)

        executable_zip = self.pre_package(project)
        zip_file.writestr("handler.zip", executable_zip.read())

        write_with_relative_path(project.root / "Makefile")

        for path in (project.root / "cmd").rglob("*"):
            if path.is_file():
                write_with_relative_path(path)

        for path in (project.root / "internal").rglob("*"):
            if path.is_file():
                write_with_relative_path(path)
