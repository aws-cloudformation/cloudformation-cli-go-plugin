## AWS CloudFormation RPDK Go Plugin

The CloudFormation Provider Development Toolkit Go Plugin allows you to autogenerate Go code based on an input schema.

This plugin library helps to provide runtime bindings for the execution of your providers by CloudFormation.

Development
-----------

For changes to the plugin, a Python virtual environment is recommended. You also need to download `cloudformation-cli` and install it first:

```
python3 -m venv env
source env/bin/activate
pip3 install cloudformation-cli
pip3 install -e .
```

Linting and running unit tests is done via [pre-commit](https://pre-commit.com/), and so is performed automatically on commit. The continuous integration also runs these checks. Manual options are available so you don't have to commit):

```
# run all hooks on all files, mirrors what the CI runs
pre-commit run --all-files
# run unit tests only. can also be used for other hooks, e.g. black, flake8, pylint-local
pre-commit run pytest-local
```

Getting started
---------------

This plugin create a sample golang project and requires golang 1.8 or above and [godeb](https://golang.github.io/dep/docs/introduction.html). For more information on installing and setting up your Go environment, please visit the offial [Golang site](https://golang.org/).


License
-------

This library is licensed under the Apache 2.0 License.
