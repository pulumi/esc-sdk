# coding: utf-8

# Copyright 2024, Pulumi Corporation.  All rights reserved.

"""
    ESC (Environments, Secrets, Config) API

    Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

    The version of the OpenAPI document: 0.1.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from setuptools import setup, find_packages  # noqa: H301

# To install the library, run the following
#
# python setup.py install
#
# prerequisite: setuptools
# http://pypi.python.org/pypi/setuptools
NAME = "pulumi-esc-sdk"
VERSION = "0.9.1.dev.0"
PYTHON_REQUIRES = ">=3.7"
REQUIRES = [
    "urllib3 >= 1.25.3, < 2.1.0",
    "python-dateutil",
    "pydantic >= 2",
    "typing-extensions >= 4.7.1",
]

setup(
    name=NAME,
    version=VERSION,
    description="ESC (Environments, Secrets, Config) API",
    author="OpenAPI Generator community",
    author_email="team@openapitools.org",
    url="",
    keywords=["OpenAPI", "OpenAPI-Generator", "ESC (Environments, Secrets, Config) API"],
    install_requires=REQUIRES,
    packages=find_packages(exclude=["test", "tests"]),
    include_package_data=True,
    license="Apache 2.0",
    long_description_content_type='text/markdown',
    long_description="""\
    Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.
    """,  # noqa: E501
    package_data={"pulumi_esc_sdk": ["py.typed"]},
)
