# coding: utf-8

"""
    ESC (Environments, Secrets, Config) API

    Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

    The version of the OpenAPI document: 0.1.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from __future__ import annotations
import pprint
import re  # noqa: F401
import json

from pydantic import BaseModel, ConfigDict, Field, StrictStr
from typing import Any, ClassVar, Dict, List, Optional
from esc.models.range import Range
from typing import Optional, Set
from typing_extensions import Self

class ExprBuiltin(BaseModel):
    """
    ExprBuiltin
    """ # noqa: E501
    name: StrictStr
    name_range: Optional[Range] = Field(default=None, alias="nameRange")
    arg_schema: Optional[Any] = Field(default=None, alias="argSchema")
    arg: Optional[Expr] = None
    __properties: ClassVar[List[str]] = ["name", "nameRange", "argSchema", "arg"]

    model_config = ConfigDict(
        populate_by_name=True,
        validate_assignment=True,
        protected_namespaces=(),
    )


    def to_str(self) -> str:
        """Returns the string representation of the model using alias"""
        return pprint.pformat(self.model_dump(by_alias=True))

    def to_json(self) -> str:
        """Returns the JSON representation of the model using alias"""
        # TODO: pydantic v2: use .model_dump_json(by_alias=True, exclude_unset=True) instead
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> Optional[Self]:
        """Create an instance of ExprBuiltin from a JSON string"""
        return cls.from_dict(json.loads(json_str))

    def to_dict(self) -> Dict[str, Any]:
        """Return the dictionary representation of the model using alias.

        This has the following differences from calling pydantic's
        `self.model_dump(by_alias=True)`:

        * `None` is only added to the output dict for nullable fields that
          were set at model initialization. Other fields with value `None`
          are ignored.
        """
        excluded_fields: Set[str] = set([
        ])

        _dict = self.model_dump(
            by_alias=True,
            exclude=excluded_fields,
            exclude_none=True,
        )
        # override the default output from pydantic by calling `to_dict()` of name_range
        if self.name_range:
            _dict['nameRange'] = self.name_range.to_dict()
        # override the default output from pydantic by calling `to_dict()` of arg
        if self.arg:
            _dict['arg'] = self.arg.to_dict()
        # set to None if arg_schema (nullable) is None
        # and model_fields_set contains the field
        if self.arg_schema is None and "arg_schema" in self.model_fields_set:
            _dict['argSchema'] = None

        return _dict

    @classmethod
    def from_dict(cls, obj: Optional[Dict[str, Any]]) -> Optional[Self]:
        """Create an instance of ExprBuiltin from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return cls.model_validate(obj)

        _obj = cls.model_validate({
            "name": obj.get("name"),
            "nameRange": Range.from_dict(obj["nameRange"]) if obj.get("nameRange") is not None else None,
            "argSchema": obj.get("argSchema"),
            "arg": Expr.from_dict(obj["arg"]) if obj.get("arg") is not None else None
        })
        return _obj

from esc.models.expr import Expr
# TODO: Rewrite to not use raise_errors
ExprBuiltin.model_rebuild(raise_errors=False)

