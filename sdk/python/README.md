# esc
Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

This Python package is automatically generated by the [OpenAPI Generator](https://openapi-generator.tech) project:

- API version: 0.1.0
- Package version: 1.0.0
- Generator version: 7.4.0
- Build package: org.openapitools.codegen.languages.PythonClientCodegen

## Requirements.

Python 3.7+

## Installation & Usage
### pip install

If the python package is hosted on a repository, you can install directly using:

```sh
pip install git+https://github.com/pulumi/esc.git
```
(you may need to run `pip` with root permission: `sudo pip install git+https://github.com/pulumi/esc.git`)

Then import the package:
```python
import esc
```

### Setuptools

Install via [Setuptools](http://pypi.python.org/pypi/setuptools).

```sh
python setup.py install --user
```
(or `sudo python setup.py install` to install the package for all users)

Then import the package:
```python
import esc
```

### Tests

Execute `pytest` to run the tests.

## Getting Started

Please follow the [installation procedure](#installation--usage) and then run the following:

```python

import esc
from esc.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to https://api.pulumi.com/api/preview
# See configuration.py for a list of all supported configuration parameters.
configuration = esc.Configuration(
    host = "https://api.pulumi.com/api/preview"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: Authorization
configuration.api_key['Authorization'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['Authorization'] = 'Bearer'


# Enter a context with an instance of the API client
with esc.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = esc.EscApi(api_client)
    org_name = 'org_name_example' # str | Organization name
    body = 'body_example' # str | Environment Yaml content

    try:
        # Checks an environment definition for errors
        api_response = api_instance.check_environment_yaml(org_name, body)
        print("The response of EscApi->check_environment_yaml:\n")
        pprint(api_response)
    except ApiException as e:
        print("Exception when calling EscApi->check_environment_yaml: %s\n" % e)

```

## Documentation for API Endpoints

All URIs are relative to *https://api.pulumi.com/api/preview*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*EscApi* | [**check_environment_yaml**](docs/EscApi.md#check_environment_yaml) | **POST** /environments/{orgName}/yaml/check | Checks an environment definition for errors
*EscApi* | [**create_environment**](docs/EscApi.md#create_environment) | **POST** /environments/{orgName}/{envName} | Create a new environment
*EscApi* | [**decrypt_environment**](docs/EscApi.md#decrypt_environment) | **GET** /environments/{orgName}/{envName}/decrypt | Reads the definition for the given environment with static secrets in plaintext
*EscApi* | [**delete_environment**](docs/EscApi.md#delete_environment) | **DELETE** /environments/{orgName}/{envName} | Delete an environment
*EscApi* | [**get_environment**](docs/EscApi.md#get_environment) | **GET** /environments/{orgName}/{envName} | Read an environment
*EscApi* | [**get_environment_e_tag**](docs/EscApi.md#get_environment_e_tag) | **HEAD** /environments/{orgName}/{envName} | Return an Environment ETag
*EscApi* | [**list_environments**](docs/EscApi.md#list_environments) | **GET** /environments/{orgName} | List environments in the organization
*EscApi* | [**open_environment**](docs/EscApi.md#open_environment) | **POST** /environments/{orgName}/{envName}/open | Open an environment session
*EscApi* | [**read_open_environment**](docs/EscApi.md#read_open_environment) | **GET** /environments/{orgName}/{envName}/open/{openSessionID} | Read an open environment
*EscApi* | [**read_open_environment_property**](docs/EscApi.md#read_open_environment_property) | **GET** /environments/{orgName}/{envName}/open//{openSessionID} | Read an open environment
*EscApi* | [**update_environment_yaml**](docs/EscApi.md#update_environment_yaml) | **PATCH** /environments/{orgName}/{envName} | Update an existing environment with Yaml file


## Documentation For Models

 - [Access](docs/Access.md)
 - [Accessor](docs/Accessor.md)
 - [CheckEnvironment](docs/CheckEnvironment.md)
 - [Environment](docs/Environment.md)
 - [EnvironmentDefinition](docs/EnvironmentDefinition.md)
 - [EnvironmentDefinitionValues](docs/EnvironmentDefinitionValues.md)
 - [EnvironmentDiagnostic](docs/EnvironmentDiagnostic.md)
 - [EnvironmentDiagnostics](docs/EnvironmentDiagnostics.md)
 - [Error](docs/Error.md)
 - [EvaluatedExecutionContext](docs/EvaluatedExecutionContext.md)
 - [Expr](docs/Expr.md)
 - [ExprBuiltin](docs/ExprBuiltin.md)
 - [Interpolation](docs/Interpolation.md)
 - [OpenEnvironment](docs/OpenEnvironment.md)
 - [OrgEnvironment](docs/OrgEnvironment.md)
 - [OrgEnvironments](docs/OrgEnvironments.md)
 - [Pos](docs/Pos.md)
 - [PropertyAccessor](docs/PropertyAccessor.md)
 - [Range](docs/Range.md)
 - [Reference](docs/Reference.md)
 - [Trace](docs/Trace.md)
 - [Value](docs/Value.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization


Authentication schemes defined for the API:
<a id="Authorization"></a>
### Authorization

- **Type**: API key
- **API key parameter name**: Authorization
- **Location**: HTTP header


## Author




