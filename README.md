## AWS CloudFormation Resource Provider Go Plugin

The CloudFormation CLI (cfn) allows you to author your own resource providers that can be used by CloudFormation.

This plugin library helps to provide Go runtime bindings for the execution of your providers by CloudFormation.

Usage
-----

If you are using this package to build resource providers for CloudFormation, install the [CloudFormation CLI Go Plugin](https://github.com/aws-cloudformation/cloudformation-cli-go-plugin) - this will automatically install the the [CloudFormation CLI](https://github.com/aws-cloudformation/cloudformation-cli)! A Python virtual environment is recommended.

```bash
pip3 install cloudformation-cli-go-plugin
```

Refer to the documentation for the [CloudFormation CLI](https://github.com/aws-cloudformation/cloudformation-cli) for usage instructions.

Development
-----------

For changes to the plugin, a Python virtual environment is recommended. Check out and install the plugin in editable mode:

```bash
python3 -m venv env
source env/bin/activate
pip3 install -e /path/to/cloudformation-cli-go-plugin
```

You may also want to check out the [CloudFormation CLI](https://github.com/aws-cloudformation/cloudformation-cli) if you wish to make edits to that. In this case, installing them in one operation works well:

```bash
pip3 install \
  -e /path/to/cloudformation-cli \
  -e /path/to/cloudformation-cli-go-plugin
```

That ensures neither is accidentally installed from PyPI.

Linting and running unit tests is done via [pre-commit](https://pre-commit.com/), and so is performed automatically on commit. The continuous integration also runs these checks. Manual options are available so you don't have to commit:

```bash
# run all hooks on all files, mirrors what the CI runs
pre-commit run --all-files
# run unit tests only. can also be used for other hooks, e.g. black, isort, pytest-local
pre-commit run pytest-local
```

Getting started
---------------

This plugin create a sample Go project and requires golang 1.8 or above and [godep](https://golang.github.io/dep/docs/introduction.html). For more information on installing and setting up your Go environment, please visit the official [Golang site](https://golang.org/).


License
-------

This library is licensed under the Apache 2.0 License.
