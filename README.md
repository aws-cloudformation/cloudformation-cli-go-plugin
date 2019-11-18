## AWS CloudFormation Resource Provider Go Plugin

The CloudFormation CLI (cfn) allows you to author your own resource providers that can be used by CloudFormation.

This plugin library helps to provide Go runtime bindings for the execution of your providers by CloudFormation.

Usage
-----

If you are using this package to build resource providers for CloudFormation, install the (CloudFormation CLI)[https://github.com/aws-cloudformation/aws-cloudformation-rpdk] and the (CloudFormation CLI Go Plugin)[https://github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin]

```
pip install cloudformation-cli
pip install cloudformation-cli-go-plugin
```

Refer to the documentation for the [CloudFormation CLI](https://github.com/aws-cloudformation/aws-cloudformation-rpdk) for usage instructions.

Development
-----------

First, you will need to install the (CloudFormation CLI)[https://github.com/aws-cloudformation/aws-cloudformation-rpdk], as it is a required dependency:

```
pip install cloudformation-cli
```

For changes to the plugin, a Python virtual environment is recommended.

```
python3 -m venv env
source env/bin/activate
# assuming cloudformation-cli has already been cloned/downloaded
pip3 install -e /path/to/aws-cloudformation-rpdk-go-plugin
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
