#!/usr/bin/env python
import os.path
import re

from setuptools import setup

HERE = os.path.abspath(os.path.dirname(__file__))


def read(*parts):
    with open(os.path.join(HERE, *parts), "r", encoding="utf-8") as fp:
        return fp.read()


# https://packaging.python.org/guides/single-sourcing-package-version/
def find_version(*file_paths):
    version_file = read(*file_paths)
    version_match = re.search(r"^__version__ = ['\"]([^'\"]*)['\"]", version_file, re.M)
    if version_match:
        return version_match.group(1)
    raise RuntimeError("Unable to find version string.")


setup(
    name="cloudformation-cli-go-plugin",
    version=find_version("python", "rpdk", "go", "__init__.py"),
    description=__doc__,
    long_description=read("README.md"),
    long_description_content_type="text/markdown",
    author="Amazon Web Services",
    author_email="aws-cloudformation-developers@amazon.com",
    url="https://github.com/aws-cloudformation/cloudformation-cli-go-plugin/",
    # https://packaging.python.org/guides/packaging-namespace-packages/
    packages=["rpdk.go"],
    package_dir={"": "python"},
    # package_data -> use MANIFEST.in instead
    include_package_data=True,
    zip_safe=True,
    install_requires=["cloudformation-cli>=0.1.2,<0.2", "semver>=2.9.0"],
    python_requires=">=3.6",
    entry_points={"rpdk.v1.languages": ["go = rpdk.go.codegen:GoLanguagePlugin"]},
    license="Apache License 2.0",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Natural Language :: English",
        "Topic :: Software Development :: Build Tools",
        "Topic :: Software Development :: Code Generators",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3 :: Only",
        "Programming Language :: Python :: 3.6",
        "Programming Language :: Python :: 3.7",
    ],
    keywords="Amazon Web Services AWS CloudFormation",
)
